package example

import (
	"errors"
	"sync"

	"glicko2/iface"

	glicko "github.com/zelenin/go-glicko2"
)

type Player struct {
	sync.RWMutex

	id      string
	isAi    bool
	aiLevel int64

	mmr float64
	RD  float64
	V   float64

	rank int
	star int

	startMatchTime  int64
	finishMatchTime int64

	*glicko.Player
}

func NewPlayer(id string, isAi bool, aiLevel int64, args iface.Args) iface.Player {
	/**
	TODO:
	算法刚启动的时候，会手动配置不同段位的补偿分数，把各个段位的分数人工区分初始化玩家的 mmr，rd 和 v，

		赛季重置也会初始化玩家的相关分数，重置方式如下；
	1. 初始评分（mmr），转换为上赛季评分*70%，最低分1000，最高分7500
	2. RD，根据当前赛季和历史最高赛季的星星数差距决定，最高700，最低为0
	3. 波动率，初始为0.06
	*/
	return &Player{
		RWMutex: sync.RWMutex{},
		id:      id,
		isAi:    isAi,
		aiLevel: aiLevel,
		mmr:     args.MMR,
		RD:      args.DR,
		V:       args.V,
		Player:  glicko.NewPlayer(glicko.NewRating(args.MMR, args.DR, args.V)),
	}
}

func (p *Player) GlickoPlayer() *glicko.Player {
	return p.Player
}

func (p *Player) IsAi() bool {
	return p.isAi
}

func (p *Player) AiLevel() int64 {
	return p.aiLevel
}

func (p *Player) GetStartMatchTimeSec() int64 {
	return p.startMatchTime
}

func (p *Player) SetStartMatchTimeSec(t int64) {
	p.startMatchTime = t
}

func (p *Player) GetFinishMatchTimeSec() int64 {
	return p.finishMatchTime
}

func (p *Player) SetFinishMatchTimeSec(t int64) {
	p.finishMatchTime = t
}

func (p *Player) ID() string {
	return p.id
}

func (p *Player) MMR() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.mmr
}

func (p *Player) Star() int {
	return p.star
}

func (p *Player) GetArgs() *iface.Args {
	p.RLock()
	defer p.RUnlock()
	/**
	TODO:
	赛季初始 5 局过后，mmr 分数开始生效；生效前沿用上赛季分数进行匹配。
	新玩家和回流玩家会进入保护期，分数不计算，对局全部按照最低分进行匹配，直到完成10次团战对局，开始使用真实分数进行匹配。
	*/
	return &iface.Args{
		MMR: p.mmr,
		DR:  p.RD,
		V:   p.V,
	}
}

func (p *Player) SetArgs(args *iface.Args) error {
	if args == nil {
		return errors.New("args is nil")
	}
	p.Lock()
	defer p.Unlock()
	p.V = args.V
	p.mmr = args.MMR
	p.RD = args.DR
	p.Player = glicko.NewPlayer(glicko.NewRating(args.MMR, args.DR, args.V))
	return nil
}

func (p *Player) ForceCancelMatch() {
	// TODO
}

func (p *Player) SetStar(star int) {
	p.star = star
}

func (p *Player) Rank() int {
	return p.rank
}

func (p *Player) SetRank(rank int) {
	p.rank = rank
}
