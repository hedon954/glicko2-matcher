# Glicko-2 Matcher
> This project implements a mathcer for glicko-2 algorithm.


## Run Example

1. Download the code and enter the root directory of the project.
2. Run matcher
   ```shell
    go test ./example -v -run Test_Matcher
   ```
3. Run settler
   ```shell
    go test ./example -v -run Test_Settler
   ```


## Hou To Use

1. Implement Player, Group, Team and Room interfaces according to your business needs.
2. Create a Macther by `NewMatcher()`, and run `matcher.Start()` to start matching.
3. When the Group starts to match, call `matcher.AddGroups(groups...)` to add the team to the matching queue and wait for the matching result.
4. When the game is over, update the `Rank` of the Team and each Player based on the result, then call `SetTler.UpdateMMR(room)`.