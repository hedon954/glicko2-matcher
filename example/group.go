package example

import (
	"fmt"
	"sync"
	"time"

	"glicko2"

	"github.com/montanaflynn/stats"
)

const (
	// 车队方差阈值
	MaliciousTeamVarianceMin  = 100000
	UnfriendlyTeamVarianceMin = 1000
)

type Group struct {
	sync.RWMutex

	id         string
	state      glicko2.GroupState
	playersMap map[string]struct{}
	players    []glicko2.Player

	startMatchTimeSec int64
}

func NewGroup(id string, players []glicko2.Player) glicko2.Group {
	g := &Group{
		RWMutex:    sync.RWMutex{},
		id:         id,
		state:      glicko2.GroupStateUnready,
		playersMap: make(map[string]struct{}),
		players:    players,
	}
	for _, p := range g.players {
		g.playersMap[p.ID()] = struct{}{}
		g.startMatchTimeSec = p.GetStartMatchTimeSec()
	}
	return g
}

func (g *Group) ID() string {
	return g.id
}

func (g *Group) GetState() glicko2.GroupState {
	g.RLock()
	defer g.RUnlock()

	return g.state
}

func (g *Group) SetState(state glicko2.GroupState) {
	g.Lock()
	defer g.Unlock()

	g.state = state
}

func (g *Group) Players() []glicko2.Player {
	g.RLock()
	defer g.RUnlock()
	return g.players
}

func (g *Group) AddPlayers(players ...glicko2.Player) {
	g.Lock()
	defer g.Unlock()

	for _, p := range players {
		_, ok := g.playersMap[p.ID()]
		if ok {
			continue
		}
		g.playersMap[p.ID()] = struct{}{}
		g.players = append(g.players, p)
	}
}

func (g *Group) RemovePlayer(player glicko2.Player) {
	g.Lock()
	defer g.Unlock()

	_, ok := g.playersMap[player.ID()]
	if !ok {
		return
	}
	delete(g.playersMap, player.ID())

	var minStartMatchTime int64
	for index, p := range g.players {
		if p == player {
			g.players = append(g.players[:index], g.players[index+1:]...)
		} else {
			if minStartMatchTime == 0 || minStartMatchTime > p.GetStartMatchTimeSec() {
				minStartMatchTime = p.GetStartMatchTimeSec()
			}
		}
	}
	g.startMatchTimeSec = minStartMatchTime
}

// AverageMMR 算出队伍的平均 mmr
func (g *Group) AverageMMR() float64 {
	total := 0.0
	for _, player := range g.players {
		total += player.MMR()
	}
	return total / float64(len(g.players))
}

// MMR 算出队伍的最大的 mmr
func (g *Group) BiggestMMR() float64 {
	mmr := 0.0
	for _, p := range g.players {
		pMMR := p.MMR()
		if pMMR > mmr {
			mmr = pMMR
		}
	}
	return mmr
}

// MMR 算出队伍的 mmr
func (g *Group) MMR() float64 {
	teamType := g.Type()
	switch teamType {
	case glicko2.GroupTypeUnfriendlyTeam:
		mmr := g.AverageMMR() * 1.5
		bMmr := g.BiggestMMR()
		if mmr > bMmr {
			mmr = bMmr
		}
		return mmr
	case glicko2.GroupTypeMaliciousTeam:
		return g.BiggestMMR()
	default:
		return g.AverageMMR()
	}
}

// Rank 队伍段位要弄平均值替代
func (g *Group) Star() int {
	if len(g.players) == 0 {
		return 0
	}
	rank := 0
	for _, p := range g.players {
		rank += p.Star()
	}
	return rank / len(g.players)
}

// Group 算出队伍的 mmr 方差
func (g *Group) MMRVariance() float64 {
	data := stats.Float64Data{}
	for _, p := range g.players {
		data = append(data, p.MMR())
	}
	variance, _ := stats.Variance(data)
	return variance
}

// Type 确定车队类型
func (g *Group) Type() glicko2.GroupType {
	if len(g.players) != 5 {
		return glicko2.GroupTypeNotTeam
	}
	variance := g.MMRVariance()
	if variance >= MaliciousTeamVarianceMin {
		return glicko2.GroupTypeMaliciousTeam
	} else if variance >= UnfriendlyTeamVarianceMin {
		return glicko2.GroupTypeUnfriendlyTeam
	} else {
		return glicko2.GroupTypeNormalTeam
	}
}

func (g *Group) CanFillAi() bool {
	// TODO: 读取配置，根据条件判断是否可以返回 ai
	now := time.Now().Unix()
	if now-g.GetStartMatchTimeSec() > 5 {
		return true
	}
	return false
}

// Print 打印 group 信息
func (g *Group) Print() {
	fmt.Printf("\t\t%s\t\t\t%d\t\t%.2f\t\t%.2f\t\t%ds\t\t\n", g.ID(), len(g.players), g.MMR(), g.AverageMMR(),
		time.Now().Unix()-g.GetStartMatchTimeSec())
}

func (g *Group) GetFinishMatchTimeSec() int64 {
	if len(g.players) == 0 {
		return 0
	}
	return g.players[0].GetFinishMatchTimeSec()
}

func (g *Group) SetFinishMatchTimeSec(t int64) {
	for _, p := range g.players {
		p.SetFinishMatchTimeSec(t)
	}
}

func (g *Group) GetStartMatchTimeSec() int64 {
	return g.startMatchTimeSec
}

func (g *Group) SetStartMatchTimeSec(t int64) {
	g.startMatchTimeSec = t
	for _, p := range g.players {
		p.SetStartMatchTimeSec(t)
	}
}
