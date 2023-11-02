package example

import (
	"fmt"
	"testing"

	"glicko2"
)

func Test_Settler(t *testing.T) {
	settler := new(glicko2.Settler)
	room := NewRoom()
	for i := 0; i < 3; i++ {
		team := NewTeam()
		team.SetRank(i + 1)
		group := NewGroup(fmt.Sprintf("team-%d-group", i+1), nil)
		group.SetState(glicko2.GroupStateQueuing)
		for j := 0; j < 5; j++ {
			player := NewPlayer(fmt.Sprintf("team-%d-player-%d", i+1, j+1), false, 0, glicko2.Args{
				MMR: 1500,
				DR:  200,
				V:   0.06,
			})
			player.SetRank(j + 1)
			group.AddPlayers(player)
		}
		team.AddGroup(group)
		room.AddTeam(team)
	}

	for i := 0; i < 10; i++ {
		settler.UpdateMMR(room)
		fmt.Println()
		fmt.Println()
	}
}
