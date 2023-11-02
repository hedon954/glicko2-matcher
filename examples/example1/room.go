package main

import (
	"strconv"

	"glicko2/iface"
)

type Room struct {
	id              int64
	teams           []iface.Team
	StartMatchTime  int64
	FinishMatchTime int64
}

func NewRoom() iface.Room {
	return &Room{
		teams: make([]iface.Team, 0, 3),
	}
}

func NewRoomWithAi(team iface.Team) iface.Room {
	newRoom := NewRoom()
	newRoom.AddTeam(team)
	// TODO: 根据实际规则填充 ai
	aiT1 := NewTeam()
	ai1G := NewGroup("ai-group-0", nil)
	for i := 0; i < TeamPlayerLimit; i++ {
		ai1G.AddPlayers(NewPlayer("ai-player-0-"+strconv.Itoa(i), true, int64(i+1), iface.Args{}))
	}
	aiT1.AddGroup(ai1G)
	aiT2 := NewTeam()
	ai2G := NewGroup("ai-group-1", nil)
	for i := 0; i < TeamPlayerLimit; i++ {
		ai2G.AddPlayers(NewPlayer("ai-player-1-"+strconv.Itoa(i), true, int64(i+1), iface.Args{}))
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

func (r *Room) Teams() []iface.Team {
	return r.teams
}

func (r *Room) AddTeam(t iface.Team) {
	r.teams = append(r.teams, t)
	tmst := t.GetStartMatchTimeSec()
	if tmst == 0 {
		return
	}
	if r.StartMatchTime == 0 || r.StartMatchTime > tmst {
		r.StartMatchTime = tmst
	}
}

func (r *Room) RemoveTeam(t iface.Team) {
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
