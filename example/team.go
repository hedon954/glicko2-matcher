package example

import (
	"sort"
	"sync"

	"github.com/hedon954/glicko2-matcher"
)

type Team struct {
	sync.RWMutex

	groups            map[string]glicko2.Group
	StartMatchTimeSec int64
	rank              int
}

func NewTeam() glicko2.Team {
	return &Team{
		RWMutex: sync.RWMutex{},
		groups:  make(map[string]glicko2.Group),
	}
}

func (t *Team) Groups() []glicko2.Group {
	res := make([]glicko2.Group, len(t.groups))
	i := 0
	for _, g := range t.groups {
		res[i] = g
		i++
	}
	return res
}

func (t *Team) AddGroup(g glicko2.Group) {
	if g.GetState() != glicko2.GroupStateQueuing {
		return
	}
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

func (t *Team) Rank() int {
	return t.rank
}

func (t *Team) SetRank(rank int) {
	t.rank = rank
}

func (t *Team) SortPlayerByRank() []glicko2.Player {
	players := make([]glicko2.Player, 0, 5)
	for _, g := range t.groups {
		players = append(players, g.Players()...)
	}
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].Rank() < players[j].Rank()
	})
	return players
}
