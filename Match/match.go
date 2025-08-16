package match

import "WitchCraft/Player"

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

type Match struct {
	ID      int
	Player1 *Player.Player
	Player2 *Player.Player
	Type    MatchType
	State   MatchState
	Turn    int
}

func New_Card(id int, player1 *Player.Player, player2 *Player.Player, TYpe MatchType, state MatchState) *Match {
	return &Match{
		ID:      id,
		Player1: player1,
		Player2: player2,
		Type:    TYpe,
		State:   state,
		Turn:    1,
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

func (m *Match) Start() {
	m.State = RUNNING
}

func (m *Match) Finish() {
	m.State = FINISHED
}

func (m *Match) NextTurn() {
	m.Turn = 3 - m.Turn
}
