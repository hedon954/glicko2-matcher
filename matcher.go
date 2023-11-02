package glicko2

import (
	"fmt"
	"sync"
	"time"

	"glicko2/iface"
)

const (
	TeamQueue   = "TeamQueue"
	NormalQueue = "NormalQueue"
)

type Matcher struct {
	quitChan chan struct{}

	NormalQueue *Queue // 普通车队
	TeamQueue   *Queue // 车队专属队列
}

// NewMatcher 是一个匹配器，包含了 TeamQueue 和 NormalQueue 两个匹配队列
func NewMatcher(
	roomChan chan iface.Room,
	queueArgs QueueArgs,
	newTeamFunc func() iface.Team,
	newRoomFunc func() iface.Room,
	newRoomWithAiFunc func(team iface.Team) iface.Room,
) *Matcher {
	return &Matcher{
		quitChan:    make(chan struct{}),
		NormalQueue: NewQueue(NormalQueue, roomChan, queueArgs, newTeamFunc, newRoomFunc, newRoomWithAiFunc),
		TeamQueue:   NewQueue(TeamQueue, roomChan, queueArgs, newTeamFunc, newRoomFunc, newRoomWithAiFunc),
	}
}

// AddGroups 添加队伍
func (qm *Matcher) AddGroups(gs ...iface.Group) {
	for _, g := range gs {
		groupType := g.Type()
		g.SetState(iface.GroupStateQueuing)
		if groupType == iface.GroupTypeNotTeam {
			qm.NormalQueue.AddGroups(g)
		} else {
			qm.TeamQueue.AddGroups(g)
		}
	}
}

func (qm *Matcher) Match() {
	ticker := time.NewTicker(time.Second).C
	for {
		select {
		case <-qm.quitChan:
			fmt.Println("\n\nGreceful exit...")
			return
		case <-ticker:
			// 取出本轮要匹配的队伍
			nGs := qm.NormalQueue.GetAndClearGroups()
			tGs := qm.TeamQueue.GetAndClearGroups()

			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				nGs = qm.NormalQueue.Match(nGs)
				wg.Done()
			}()
			go func() {
				tGs = qm.TeamQueue.Match(tGs)
				wg.Done()
			}()
			wg.Wait()

			// 判断哪些 group 需要从专属队列从移动到普通队列
			now := time.Now()
			for _, g := range tGs {
				needMove := false
				matchTime := now.Unix() - g.GetStartMatchTimeSec()
				switch g.Type() {
				case iface.GroupTypeMaliciousTeam:
					if matchTime >= qm.TeamQueue.MaliciousTeamWaitTimeSec {
						needMove = true
					}
				case iface.GroupTypeUnfriendlyTeam:
					if matchTime >= qm.TeamQueue.UnfriendlyTeamWaitTimeSec {
						needMove = true
					}
				case iface.GroupTypeNormalTeam:
					if matchTime >= qm.TeamQueue.NormalTeamWaitTimeSec {
						needMove = true
					}
				}
				if needMove {
					qm.NormalQueue.AddGroups(g)
				} else {
					qm.TeamQueue.AddGroups(g)
				}
			}

			// 将普通队列中上轮没成功匹配的加回去，下轮重新匹配
			qm.NormalQueue.AddGroups(nGs...)

			fmt.Println("QueueName\t\tTmpTeam\t\tTmpRoom\t\tGroup\t\t")
			fmt.Printf("%s\t\t%d\t\t%d\t\t%d\t\t\n", qm.NormalQueue.Name, len(qm.NormalQueue.tmpTeam),
				len(qm.NormalQueue.tmpRoom), len(qm.NormalQueue.Groups))
			fmt.Printf("%s\t\t%d\t\t%d\t\t%d\t\t\n", qm.TeamQueue.Name, len(qm.TeamQueue.tmpTeam),
				len(qm.TeamQueue.tmpRoom), len(qm.TeamQueue.Groups))
			fmt.Println()
		}
	}
}

func (qm *Matcher) Stop() {
	qm.NormalQueue.stopMatch()
	qm.TeamQueue.stopMatch()
	qm.quitChan <- struct{}{}
}
