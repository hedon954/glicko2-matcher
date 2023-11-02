package glicko2

// Team 是一个阵营的抽象，由 1~n 个 Group 组成
type Team interface {
	Groups() []Group

	// 添加 group 到 team 中
	AddGroup(group Group)

	// 移除 group
	RemoveGroup(groupId string)

	// 玩家数量
	PlayerCount() int

	// 阵营的平均 mmr
	AverageMMR() float64

	// 阵营的段位值
	Star() int

	// 阵营的开始匹配时间，取最早的玩家的
	GetStartMatchTimeSec() int64

	// 阵营的完成匹配时间
	GetFinishMatchTimeSec() int64
	SetFinishMatchTimeSec(t int64)

	// 当前阵营是否是 ai
	IsAi() bool

	// 赛后在房间内的排名
	Rank() int
	SetRank(rank int)

	// 赛后根据排名获取玩家列表
	SortPlayerByRank() []Player
}
