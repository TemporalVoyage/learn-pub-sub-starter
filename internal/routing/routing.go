package routing

const (
	ArmyMovesPrefix = "army_moves"

	WarRecognitionsPrefix = "war"

	PauseKey = "pause"

	GameLogSlug = "game_logs"
)

type Acktype int
type SimpleQueueType int

const (
	ExchangePerilDirect = "peril_direct"
	ExchangePerilTopic  = "peril_topic"
)
const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)
const (
	Ack Acktype = iota
	NackDis
	NackRe
)
