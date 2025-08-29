package match

import (
	"WitchCraft/Cards"
	"WitchCraft/Player"
	"encoding/json"
)

const (
	NORMAL = iota
	RANKED
)

const (
	WAITING = iota
	RUNNING
	FINISHED
)

type MatchState uint8

type MatchType uint8

type PlayedCard struct {
	Card     Cards.Card
	Atribute string
}

type Match_Message struct {
	PlayerId int             `json:"id"`
	Action   string          `json:"action"`
	Data     json.RawMessage `json:"data"`
}

type Match struct {
	ID      int
	Player1 *Player.Player
	Player2 *Player.Player
	Type    MatchType
	State   MatchState
	Turn    int

	PlayedCard1 *PlayedCard
	PlayedCard2 *PlayedCard
	Round       int

	MatchChan chan Match_Message
}

func New_match(id int, player1 *Player.Player, player2 *Player.Player, TYpe MatchType, state MatchState, turn int) *Match {
	return &Match{
		ID:        id,
		Player1:   player1,
		Player2:   player2,
		Type:      TYpe,
		State:     state,
		Turn:      turn,
		MatchChan: make(chan Match_Message, 1),
	}
}

func (ms MatchState) String() string {
	switch ms {
	case WAITING:
		return "Waiting"
	case RUNNING:
		return "Running"
	case FINISHED:
		return "FINISHED"
	default:
		return "unknown"
	}
}

func (mt MatchType) String() string {
	switch mt {
	case NORMAL:
		return "Normal Game"
	case RANKED:
		return "Ranked Game"
	default:
		return "unknown"
	}
}
