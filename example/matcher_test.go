package example

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hedon954/glicko2-matcher"
)

func Test_Matcher(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var roomId = atomic.Int64{}

	roomChan := make(chan glicko2.Room, 128)

	queueArgs := glicko2.QueueArgs{
		RoomPlayerLimit:           RoomPlayerLimit,
		TeamPlayerLimit:           TeamPlayerLimit,
		RoomTeamLimit:             RoomTeamLimit,
		NormalTeamWaitTimeSec:     NormalTeamWaitTimeSec,
		UnfriendlyTeamWaitTimeSec: UnfriendlyTeamWaitTimeSec,
		MaliciousTeamWaitTimeSec:  MaliciousTeamWaitTimeSec,
		MatchRanges: []glicko2.MatchRange{
			{
				MaxMatchSec:   1,
				MMRGapPercent: 10,
				CanJoinTeam:   false,
				RankGap:       0,
			},
			{
				MaxMatchSec:   5,
				MMRGapPercent: 20,
				CanJoinTeam:   false,
				RankGap:       0,
			},
			{
				MaxMatchSec:   10,
				MMRGapPercent: 30,
				CanJoinTeam:   true,
				RankGap:       0,
			},
			{
				MaxMatchSec:   30,
				MMRGapPercent: 0,
				CanJoinTeam:   true,
				RankGap:       0,
			},
		},
	}

	qm := glicko2.NewMatcher(roomChan, queueArgs, NewTeam, NewRoom, NewRoomWithAi)

	// 异步随机生成 group
	go func() {
		for i := 0; i < 100; i++ {
			var players []glicko2.Player
			count := rand.Intn(5) + 1
			for j := 0; j < count; j++ {
				p := NewPlayer("", false, 0,
					glicko2.Args{
						MMR: float64(rand.Intn(3000)),
						DR:  0,
						V:   0,
					})
				players = append(players, p)
			}
			newGroup := NewGroup(fmt.Sprintf("Group%d", i+1), players)
			qm.AddGroups(newGroup)
			ssec := rand.Intn(200)
			time.Sleep(time.Duration(ssec) * time.Millisecond)
		}
	}()

	// 异步启动匹配
	go qm.Match()

	// 进程退出
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	for {
		select {
		// 模拟消费 room
		case tr := <-roomChan:
			now := time.Now().Unix()
			rId := roomId.Add(1)
			fmt.Println("-------------------------------------------------------------------")
			fmt.Printf("| Room[%d] Match successful, cast time %ds, hasAi: %t\n", rId,
				now-tr.GetStartMatchTimeSec(), tr.HasAi())
			for j, team := range tr.Teams() {
				fmt.Printf("|   Team %d average mmr: %.2f, isAi: %t, cost time %ds\n", j+1,
					team.AverageMMR(), team.IsAi(), now-team.GetStartMatchTimeSec())
				for _, group := range team.Groups() {
					group.SetState(glicko2.GroupStateMatched)
					fmt.Printf("|     %s mmr: %.2f, player count: %d, team type: %d, cost time %ds\n", group.ID(),
						group.MMR(),
						len(group.Players()), group.Type(),
						now-group.GetStartMatchTimeSec())
				}
			}
			fmt.Println("-------------------------------------------------------------------")
			fmt.Println()
		case <-ch:
			gs1, gs2 := qm.Stop()

			fmt.Println()
			fmt.Println()
			fmt.Println("--------------- finish --------------")

			fmt.Println("normal queue left group count:", len(gs1))
			fmt.Printf("\t\tGroupId\t\t\tPlayerCount\t\tmmr\t\tAvgMMR\t\tMatchTime\t\t\n")
			for _, g := range gs1 {
				g.Print()
			}
			fmt.Println()
			fmt.Println("team queue left group count:", len(gs2))
			fmt.Printf("\t\tGroupId\t\t\tPlayerCount\t\tmmr\t\tAvgMMR\t\tMatchTime\t\t\n")
			for _, g := range gs2 {
				g.Print()
			}
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
