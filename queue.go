package glicko2

import (
	"math"
	"sort"
	"sync"
	"time"

	"glicko2/iface"
)

const (
	// 每 5 轮会将 tmpRoom 和 tmpTeam 清空并打散重新匹配
	refreshTurn = 5
)

// Queue 是一个匹配队列
type Queue struct {
	sync.Mutex
	Name          string                           // 队列名称
	Groups        []iface.Group                    // 在队列中的队伍，对于 Groups 的所有处理都要加锁
	tmpTeam       []iface.Team                     // 匹配过程中的临时阵营，每 5 轮匹配后会打散重来，只能在 Match 中调用，不可以并发调用
	tmpRoom       []iface.Room                     // 匹配过程中的临时房间，每 5 轮匹配后会打散重来，只能在 Match 中调用，不可以并发调用
	roomChan      chan iface.Room                  // 匹配成功的房间会投进这个 channel
	newTeam       func() iface.Team                // 构建新 team 的方法
	newRoom       func() iface.Room                // 构建新 room 的方法
	newRoomWithAi func(team iface.Team) iface.Room // 构建带 ai 的新 room 的方法
	matchTurn     int                              // 匹配轮次，对 5 取模

	QueueArgs
}

type QueueArgs struct {
	RoomPlayerLimit int // 房间总人数上线
	TeamPlayerLimit int // 阵营总人数上限
	RoomTeamLimit   int // 房间总阵营数

	NormalTeamWaitTimeSec     int64 // 普通车队在专属队列中的匹配时长
	UnfriendlyTeamWaitTimeSec int64 // 不友好车队在匹配队列中的匹配时长
	MaliciousTeamWaitTimeSec  int64 // 恶意车队在匹配队列中的匹配时长

	MatchRanges []MatchRange // 匹配范围策略
}

type MatchRange struct {
	MaxMatchSec   int64 // 最长匹配时间s（不包含）
	MMRGapPercent int   // 允许的 mmr 差距百分比(0~100)（包含），0 表示无限制
	CanJoinTeam   bool  // 是否加入 5 人车队
	RankGap       int   // 允许的段位差距数（包含），0 表示无限制
}

var defaultMatchRange = MatchRange{
	MaxMatchSec:   15,
	MMRGapPercent: 10,
	CanJoinTeam:   false,
	RankGap:       12,
}

func NewQueue(
	name string, roomChan chan iface.Room, args QueueArgs, newTeamFunc func() iface.Team,
	newRoomFunc func() iface.Room,
	newRoomWithAiFunc func(team iface.Team) iface.Room,
) *Queue {
	return &Queue{
		Mutex:         sync.Mutex{},
		Name:          name,
		roomChan:      roomChan,
		Groups:        make([]iface.Group, 0, 128),
		tmpTeam:       make([]iface.Team, 0, 128),
		tmpRoom:       make([]iface.Room, 0, 128),
		newTeam:       newTeamFunc,
		newRoom:       newRoomFunc,
		newRoomWithAi: newRoomWithAiFunc,
		QueueArgs:     args,
	}
}

func (q *Queue) SortedGroups() []iface.Group {
	q.Lock()
	defer q.Unlock()

	sort.Slice(q.Groups, func(i, j int) bool {
		return q.Groups[i].MMR() < q.Groups[j].MMR()
	})
	return q.Groups
}

func (q *Queue) AllGroups() []iface.Group {
	q.Lock()
	defer q.Unlock()

	return q.Groups
}

func (q *Queue) AddGroups(gs ...iface.Group) {
	q.Lock()
	defer q.Unlock()

	for _, g := range gs {
		if g.GetStartMatchTimeSec() == 0 {
			g.SetStartMatchTimeSec(time.Now().Unix())
		}
	}
	q.Groups = append(q.Groups, gs...)
}

// GetAndClearGroups 取出要匹配的 group 并且清空当前 groups 列表
func (q *Queue) GetAndClearGroups() []iface.Group {
	q.Lock()
	defer q.Unlock()

	res := make([]iface.Group, 0, len(q.Groups))
	for _, g := range q.Groups {
		if g.GetState() == iface.GroupStateQueuing {
			res = append(res, g)
		}
	}
	q.Groups = make([]iface.Group, 0, 128)
	return res
}

// clearTmp 清除 tmpRoom 和 tmpTeam 并归位 groups
func (q *Queue) clearTmp() []iface.Group {
	groups := make([]iface.Group, 0, 128)
	for _, t := range q.tmpTeam {
		groups = append(groups, t.Groups()...)
	}
	for _, r := range q.tmpRoom {
		for _, rt := range r.Teams() {
			groups = append(groups, rt.Groups()...)
		}
	}
	q.tmpTeam = make([]iface.Team, 0, 128)
	q.tmpRoom = make([]iface.Room, 0, 128)
	return groups
}

// Match 队列匹配逻辑
func (q *Queue) Match(groups []iface.Group) []iface.Group {
	var tmpTeam = q.tmpTeam
	var tmpRoom = q.tmpRoom

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].MMR() < groups[j].MMR()
	})

	// 获取所有的玩家个数
	totalPlayerCount := 0
	for _, g := range groups {
		totalPlayerCount += len(g.Players())
	}
	for _, t := range tmpTeam {
		totalPlayerCount += t.PlayerCount()
	}
	for _, r := range tmpRoom {
		totalPlayerCount += r.PlayerCount()
	}

	sort.Slice(tmpTeam, func(i, j int) bool {
		return tmpTeam[i].AverageMMR() < tmpTeam[j].AverageMMR()
	})

	// 尝试构建 totalPlayerCount/RoomPlayerLimit + 1 个 room
	for k := 0; k < totalPlayerCount/q.RoomPlayerLimit+1; k++ {
		// 优先把 tmp team 填满
		for _, tt := range tmpTeam {
			for tt.PlayerCount() != q.TeamPlayerLimit {
				var found bool
				groups, found = q.findGroupForTeam(tt, groups)
				if !found {
					break
				}
			}
		}

		// 获取还在队列中的玩家数，尝试构建新的 team
		notInTeamPlayerCount := 0
		for _, g := range groups {
			notInTeamPlayerCount += len(g.Players())
		}

		// 再去构建新的 team
		for i := 0; i < notInTeamPlayerCount/q.TeamPlayerLimit+1; i++ {
			team := q.newTeam()
			for team.PlayerCount() != q.TeamPlayerLimit {
				var found bool
				groups, found = q.findGroupForTeam(team, groups)
				if !found {
					break
				}
			}
			if team.PlayerCount() == 0 {
				break
			}
			tmpTeam = append(tmpTeam, team)
		}

		// 优先在 tmpRoom 中创建房间
		for _, tr := range tmpRoom {
			if len(tr.Teams()) == q.RoomTeamLimit {
				continue
			}
			for len(tr.Teams()) != q.RoomTeamLimit {
				var found bool
				tmpTeam, found = q.findTeamForRoom(tr, tmpTeam)
				if !found {
					break
				}
			}
		}

		// 尝试继续创建新的房间
		tryRoomTimes := len(tmpTeam) / q.RoomTeamLimit
		for l := 0; l < tryRoomTimes+1; l++ {
			room := q.newRoom()
			for len(room.Teams()) != q.RoomTeamLimit {
				var found bool
				tmpTeam, found = q.findTeamForRoom(room, tmpTeam)
				if !found {
					break
				}
			}
			if len(room.Teams()) == 0 {
				break
			}
			tmpRoom = append(tmpRoom, room)
		}

		// 尝试填充 ai
		for _, tr := range tmpRoom {
			teams := tr.Teams()
			if len(teams) == 0 || len(teams) == q.RoomTeamLimit {
				continue
			}
			for _, team := range teams {
				canFillAi := true
				for _, g := range tr.Teams()[0].Groups() {
					if !g.CanFillAi() {
						canFillAi = false
						break
					}
				}
				if !canFillAi {
					continue
				}
				tr.RemoveTeam(team)
				newRoom := q.newRoomWithAi(team)
				tmpRoom = append(tmpRoom, newRoom)
			}
		}

		// 整理房间信息
		newTmpRoom := make([]iface.Room, 0)
		for _, tr := range tmpRoom {
			if len(tr.Teams()) == q.RoomTeamLimit {
				now := time.Now().Unix()
				tr.SetFinishMatchTimeSec(now)
				go func(room iface.Room) {
					q.roomChan <- room
				}(tr)
				continue
			}
			newTmpRoom = append(newTmpRoom, tr)
		}
		tmpRoom = newTmpRoom
	}

	// 每 refreshTurn 轮都打散重来
	q.matchTurn = (q.matchTurn + 1) % refreshTurn
	if q.matchTurn == 0 {
		gs := q.clearTmp()
		groups = append(groups, gs...)
	} else {
		q.tmpTeam = tmpTeam
		q.tmpRoom = tmpRoom
	}

	// TODO: 谨慎考虑匹配一般玩家取消匹配的参加
	return groups
}

// findGroupForTeam 从 groups 中找到适合 team 的 group 并加入其中
func (q *Queue) findGroupForTeam(team iface.Team, groups []iface.Group) ([]iface.Group, bool) {
	// 第1个队伍直接进
	if team.PlayerCount() == 0 && len(groups) > 0 {
		team.AddGroup(groups[0])
		groups = groups[1:]
		return groups, true
	}

	// 寻找平均 mmr 最接近的 group 组成一个 team
	closestIndex := -1
	for i, group := range groups {
		// 优先找能凑满队的
		if team.PlayerCount()+len(group.Players()) == q.TeamPlayerLimit && (closestIndex == -1 || math.Abs(group.MMR()-team.AverageMMR()) < math.Abs(groups[closestIndex].MMR()-team.AverageMMR())) {
			closestIndex = i
		}
	}
	if closestIndex == -1 {
		// 不能一次性组满队，就先临时组一个队，后面再尝试组满
		for i, group := range groups {
			if team.PlayerCount()+len(group.Players()) <= q.TeamPlayerLimit && (closestIndex == -1 || math.Abs(group.MMR()-team.AverageMMR()) < math.Abs(groups[closestIndex].MMR()-team.AverageMMR())) {
				closestIndex = i
			}
		}
		// 如果没有找到合适的 group，则直接返回，这里一般是因为 group 列表为空
		if closestIndex == -1 {
			return groups, false
		}
	}

	if q.canGroupTogether(team, groups[closestIndex]) {
		team.AddGroup(groups[closestIndex])
		groups = append(groups[:closestIndex], groups[closestIndex+1:]...)
		return groups, true
	}

	return groups, false
}

// findTeamForRoom 从 tmpTeam 中找到合适 room 的 team 并加入其中
func (q *Queue) findTeamForRoom(room iface.Room, tmpTeam []iface.Team) ([]iface.Team, bool) {
	for tPos, tt := range tmpTeam {
		if len(room.Teams()) >= q.RoomTeamLimit {
			break
		}
		// 只有当 team 已经组建完毕了，才可以加入到 room 中
		if tt.PlayerCount() != q.TeamPlayerLimit {
			continue
		}
		// 如果 room 中没有 team，则第 1 个直接加入 room 中
		if len(room.Teams()) == 0 {
			room.AddTeam(tt)
			tmpTeam = append(tmpTeam[:tPos], tmpTeam[tPos+1:]...)
			return tmpTeam, true
		} else {
			if q.canTeamTogether(room, tt) {
				room.AddTeam(tt)
				tmpTeam = append(tmpTeam[:tPos], tmpTeam[tPos+1:]...)
				return tmpTeam, true
			}
		}
	}

	// 没找着
	return tmpTeam, false
}

// canGroupTogether 判断队伍之间是否可以组成一个阵营
func (q *Queue) canGroupTogether(team iface.Team, group iface.Group) bool {
	for _, g := range team.Groups() {
		mr := q.getMatchRange(g.GetStartMatchTimeSec(), group.GetStartMatchTimeSec())

		// 是否加入车队
		if len(g.Players()) != q.TeamPlayerLimit && !mr.CanJoinTeam && len(group.Players()) == q.TeamPlayerLimit {
			return false
		}

		// mmr 是否匹配
		gMMR := g.MMR()
		if mr.MMRGapPercent != 0 && math.Abs(gMMR-group.MMR()) > g.MMR()*float64(mr.MMRGapPercent)/100 {
			return false
		}

		// 段位是否匹配
		if mr.RankGap != 0 && int(math.Abs(float64(g.Star()-group.Star()))) > mr.RankGap {
			return false
		}
	}
	return true
}

// canTeamTogether 判断阵营之间是否可以组成一个房间
func (q *Queue) canTeamTogether(room iface.Room, tt iface.Team) bool {
	// 判断 tt 是否满足跟当前 room 中的所有 team 匹配的条件
	// 只要有一个不满足，就返回 false
	for _, t := range room.Teams() {
		mr := q.getMatchRange(t.GetStartMatchTimeSec(), tt.GetStartMatchTimeSec())
		// 是否加入车队
		if len(t.Groups()) > 1 && !mr.CanJoinTeam && len(tt.Groups()) == 1 {
			return false
		}

		// mmr 是否匹配
		tMMR := t.AverageMMR()
		if mr.MMRGapPercent != 0 && math.Abs(tMMR-tt.AverageMMR()) > tMMR*float64(mr.MMRGapPercent)/100 {
			return false
		}

		// 段位是否匹配
		if mr.RankGap != 0 && int(math.Abs(float64(t.Star()-tt.Star()))) > mr.RankGap {
			return false
		}
	}
	return true
}

// getMatchRange 获取匹配范围
func (q *Queue) getMatchRange(mst1, mst2 int64) MatchRange {
	if len(q.MatchRanges) == 0 {
		return defaultMatchRange
	}

	// 以匹配时间短的那个为准
	mt := int64(math.Min(float64(mst1), float64(mst2)))
	for _, mr := range q.MatchRanges {
		if mt < mr.MaxMatchSec {
			return mr
		}
	}

	// 默认返回最后一个
	return q.MatchRanges[len(q.MatchRanges)-1]
}

// stopMatch 取消匹配
func (q *Queue) stopMatch() {
	q.Lock()
	defer q.Unlock()

	groups := q.clearTmp()
	q.Groups = append(q.Groups, groups...)
	for _, g := range q.Groups {
		if g.GetState() != iface.GroupStateQueuing {
			continue
		}
		for _, p := range g.Players() {
			if !p.IsAi() {
				p.ForceCancelMatch()
			}
		}
		g.SetState(iface.GroupStateUnready)
	}
	q.Groups = make([]iface.Group, 0)
}
