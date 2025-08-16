package match

import (
	"WitchCraft/Player"
	"sync"
)

type Match_Manager struct {
	mu      sync.Mutex
	Matches []*Match
}

var nextID int
var muID sync.Mutex

func generateID() int {
	muID.Lock()
	defer muID.Unlock()
	nextID++
	return nextID
}

func NewMatchManager() *Match_Manager {
	return &Match_Manager{
		Matches: make([]*Match, 0),
	}
}

func (m *Match_Manager) CreateMatch(player1 *Player.Player, player2 *Player.Player, TYpe MatchType, state MatchState) *Match {
	m.mu.Lock()
	defer m.mu.Unlock()

	match := New_match(generateID(), player1, player2, TYpe, state)
	m.Matches = append(m.Matches, match)

	return match

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
