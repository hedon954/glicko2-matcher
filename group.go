package glicko2

// GroupState队伍状态
type GroupState uint8

const (
	GroupStateUnready GroupState = iota // 未准备
	GroupStateQueuing                   // 匹配中
	GroupStateMatched                   // 匹配完成
)

// GroupType 车队类型
type GroupType uint8

const (
	// 车队类型
	GroupTypeNotTeam GroupType = iota
	GroupTypeNormalTeam
	GroupTypeUnfriendlyTeam
	GroupTypeMaliciousTeam
)

// Group 是一个队伍，
// 玩家可以自行组队，单个玩家开始匹配的时候也会为其单独创建一个队伍，
// 匹配前后队伍都不会被拆开。
type Group interface {

	// 队伍 ID
	ID() string

	// 获取队伍里的玩家列表
	Players() []Player

	// 添加玩家到队伍中
	AddPlayers(players ...Player)

	// 从队伍中移除玩家
	RemovePlayer(player Player)

	// 获取队伍 mmr 值
	MMR() float64

	// 获取队伍段位值
	Star() int

	// 获取队伍的 mmr 方差
	MMRVariance() float64

	// 获取队伍的平均 mmr 值
	AverageMMR() float64

	// 获取队伍中最大的 mmr 值
	BiggestMMR() float64

	// 获取队伍状态
	GetState() GroupState

	// 更新队伍状态
	SetState(state GroupState)

	// 开始匹配的时间，取 player 中最早的
	GetStartMatchTimeSec() int64
	SetStartMatchTimeSec(t int64)

	// 结束匹配的时间
	GetFinishMatchTimeSec() int64
	SetFinishMatchTimeSec(t int64)

	// 获取车队类型
	Type() GroupType

	// 当返回 true 时，会自动填充 Ai 组成房间，第二个返回值为填充的 ai team
	CanFillAi() bool

	// 打印信息
	Print()
}
