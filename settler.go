package glicko2

import (
	"fmt"

	glicko "github.com/zelenin/go-glicko2"
)

// Settler 游戏结算器
type Settler struct{}

func (s *Settler) UpdateMMR(room Room) {

	// 开启一个 glicko-2 计算周期
	period := glicko.NewRatingPeriod()

	// 阵营间
	// T1 > T2 > T3
	// T2 > T3
	teams := room.SortTeamByRank()
	for i := 0; i < len(teams)-1; i++ {
		team1 := teams[i]
		for j := i + 1; j < len(teams); j++ {
			teamj := teams[j]
			for _, g := range team1.Groups() {
				for _, team1Player := range g.Players() {
					if team1Player.IsAi() {
						continue
					}
					t1p := team1Player.GlickoPlayer()
					for _, g2 := range teamj.Groups() {
						for _, teamjPlayer := range g2.Players() {
							if teamjPlayer.IsAi() {
								continue
							}
							period.AddMatch(t1p, teamjPlayer.GlickoPlayer(),
								glicko.MATCH_RESULT_WIN)
						}
					}
				}
			}
		}
	}

	// 阵营内
	// P1 > P2 > P3 > P4 > P5
	// P2 > P3 > P4 > P5
	// P3 > P4 > P5
	// P4 > P5
	for _, team1 := range teams {
		team1Players := team1.SortPlayerByRank()
		for j := 0; j < len(team1Players)-1; j++ {
			if team1Players[j].IsAi() {
				continue
			}
			tp1 := team1Players[j].GlickoPlayer()
			for k := j + 1; k < len(team1Players); k++ {
				if team1Players[k].IsAi() {
					continue
				}
				period.AddMatch(tp1, team1Players[k].GlickoPlayer(), glicko.MATCH_RESULT_WIN)
			}
		}
	}

	// 计算结果
	period.Calculate()

	// 输出更新后的结果
	for _, team := range teams {
		players := team.SortPlayerByRank()
		for i := 0; i < len(players); i++ {
			if players[i].IsAi() {
				continue
			}
			rating := players[i].GlickoPlayer().Rating()
			_ = players[i].SetArgs(&Args{
				MMR: rating.R(),
				DR:  rating.Rd(),
				V:   rating.Sigma(),
			})
			fmt.Printf("Player #%s mmr: %0.2f, rd: %0.2f, v: %0.2f\n", players[i].ID(), rating.R(), rating.Rd(),
				rating.Sigma())
		}
	}

	fmt.Println("-----------------------------")
	fmt.Println()
}
