package example

import (
	"sort"
	"strconv"

	"glicko2"
)

const (
	// 车队在专属队列中的匹配时长
	NormalTeamWaitTimeSec     int64 = 5
	UnfriendlyTeamWaitTimeSec int64 = 10
	MaliciousTeamWaitTimeSec  int64 = 15

	RoomPlayerLimit = 15 // 房间总人数
	TeamPlayerLimit = 5  // 阵营总人数
	RoomTeamLimit   = 3  // 房间总阵营数
)

type Room struct {
	id              int64
	teams           []glicko2.Team
	StartMatchTime  int64
	FinishMatchTime int64
}

func NewRoom() glicko2.Room {
	return &Room{
		teams: make([]glicko2.Team, 0, 3),
	}
}

func NewRoomWithAi(team glicko2.Team) glicko2.Room {
	newRoom := NewRoom()
	newRoom.AddTeam(team)
	// TODO: 根据实际规则填充 ai
	aiT1 := NewTeam()
	ai1G := NewGroup("ai-group-0", nil)
	for i := 0; i < TeamPlayerLimit; i++ {
		ai1G.AddPlayers(NewPlayer("ai-player-0-"+strconv.Itoa(i), true, int64(i+1), glicko2.Args{}))
	}
	aiT1.AddGroup(ai1G)
	aiT2 := NewTeam()
	ai2G := NewGroup("ai-group-1", nil)
	for i := 0; i < TeamPlayerLimit; i++ {
		ai2G.AddPlayers(NewPlayer("ai-player-1-"+strconv.Itoa(i), true, int64(i+1), glicko2.Args{}))
	}
	aiT2.AddGroup(ai2G)
	newRoom.AddTeam(aiT1)
	newRoom.AddTeam(aiT2)
	return newRoom
}

func (r *Room) GetID() int64 {
	return r.id
}

func (r *Room) SetID(rid int64) {
	r.id = rid
}

func (r *Room) Teams() []glicko2.Team {
	return r.teams
}

func (r *Room) AddTeam(t glicko2.Team) {
	r.teams = append(r.teams, t)
	tmst := t.GetStartMatchTimeSec()
	if tmst == 0 {
		return
	}
	if r.StartMatchTime == 0 || r.StartMatchTime > tmst {
		r.StartMatchTime = tmst
	}
}

func (r *Room) RemoveTeam(t glicko2.Team) {
	for i, rt := range r.teams {
		if rt == t {
			r.teams = append(r.teams[:i], r.teams[i+1:]...)
			break
		}
	}
	return
}

func (r *Room) GetStartMatchTimeSec() int64 {
	return r.StartMatchTime
}

func (r *Room) GetFinishMatchTimeSec() int64 {
	return r.FinishMatchTime
}

func (r *Room) PlayerCount() int {
	count := 0
	for _, t := range r.teams {
		count += t.PlayerCount()
	}
	return count
}

func (r *Room) SetFinishMatchTimeSec(t int64) {
	for _, team := range r.teams {
		team.SetFinishMatchTimeSec(t)
	}
	r.FinishMatchTime = t
}

func (r *Room) HasAi() bool {
	for _, t := range r.teams {
		if t.IsAi() {
			return true
		}
	}
	return false
}

func (r *Room) SortTeamByRank() []glicko2.Team {
	sort.SliceStable(r.teams, func(i, j int) bool {
		return r.teams[i].Rank() < r.teams[j].Rank()
	})
	return r.teams
}
