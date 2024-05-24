package main

import (
	"fmt"

	"github.com/temporalvoyage/learn-pub-sub-starter/internal/gamelogic"
	"github.com/temporalvoyage/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(gl routing.GameLog) routing.Acktype {
	return func(gl routing.GameLog) routing.Acktype {
		fmt.Println("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			return routing.NackRe
		}
		return routing.Ack
	}
}
