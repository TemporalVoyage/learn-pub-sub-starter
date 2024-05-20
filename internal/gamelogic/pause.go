package gamelogic

import (
	"fmt"

	"github.com/temporalvoyage/learn-pub-sub-starter/internal/routing"
)

func (gs *GameState) HandlePause(ps routing.PlayingState) int {
	defer fmt.Println("------------------------")
	fmt.Println()
	if ps.IsPaused {
		fmt.Println("==== Pause Detected ====")
		gs.pauseGame()
	} else {
		fmt.Println("==== Resume Detected ====")
		gs.resumeGame()
	}
	return routing.Ack
}
