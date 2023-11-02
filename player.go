package glicko2

import (
	glicko "github.com/zelenin/go-glicko2"
)

// Player 是一个玩家的抽象
type Player interface {

	// 玩家ID
	ID() string

	// 是否是 AI
	IsAi() bool

	// AI 难度等级
	AiLevel() int64

	// 获取 mmr 值
	MMR() float64

	// 段位值
	Star() int
	SetStar(star int)

	// 获取参数
	GetArgs() *Args

	// 更新参数
	SetArgs(args *Args) error

	// 开始匹配的时间
	GetStartMatchTimeSec() int64
	SetStartMatchTimeSec(t int64)

	// 结束匹配的时间
	GetFinishMatchTimeSec() int64
	SetFinishMatchTimeSec(t int64)

	// 赛后在阵营内的排名
	Rank() int
	SetRank(rank int)

	// 强制退出时对每个玩家的处理逻辑
	ForceCancelMatch()

	// glicko-2 算法的玩家抽象示例
	GlickoPlayer() *glicko.Player
}
