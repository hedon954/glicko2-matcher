package main

import (
	"sync"

	"glicko2/iface"
)

type Team struct {
	sync.RWMutex

	groups            map[string]iface.Group
	StartMatchTimeSec int64
}

func NewTeam() iface.Team {
	return &Team{
		RWMutex: sync.RWMutex{},
		groups:  make(map[string]iface.Group),
	}
}

func (t *Team) Groups() []iface.Group {
	res := make([]iface.Group, len(t.groups))
	i := 0
	for _, g := range t.groups {
		res[i] = g
		i++
	}
	return res
}

func (t *Team) AddGroup(g iface.Group) {
	t.groups[g.ID()] = g
	gmst := g.GetStartMatchTimeSec()
	if gmst == 0 {
		return
	}
	if t.StartMatchTimeSec == 0 || t.StartMatchTimeSec > gmst {
		t.StartMatchTimeSec = gmst
	}
}

func (t *Team) RemoveGroup(groupId string) {
	delete(t.groups, groupId)
}

func (t *Team) PlayerCount() int {
	count := 0
	for _, group := range t.groups {
		count += len(group.Players())
	}
	return count
}

func (t *Team) AverageMMR() float64 {
	if len(t.groups) == 0 {
		return 0
	}
	total := 0.0
	for _, group := range t.groups {
		total += group.MMR()
	}
	return total / float64(len(t.groups))
}

func (t *Team) Star() int {
	if len(t.groups) == 0 {
		return 0
	}
	rank := 0
	for _, g := range t.groups {
		rank += g.Star()
	}
	return rank / len(t.groups)
}

func (t *Team) SetFinishMatchTimeSec(t2 int64) {
	for _, g := range t.groups {
		g.SetFinishMatchTimeSec(t2)
	}
}

func (t *Team) GetStartMatchTimeSec() int64 {
	return t.StartMatchTimeSec
}

func (t *Team) GetFinishMatchTimeSec() int64 {
	for _, g := range t.groups {
		return g.GetFinishMatchTimeSec()
	}
	return 0
}

func (t *Team) IsAi() bool {
	for _, g := range t.groups {
		for _, p := range g.Players() {
			if p.IsAi() {
				return true
			}
		}
	}
	return false
}
