package match

import (
	"WitchCraft/Player"
	"errors"
	"sync"
)

type Queue []Player.Player

type Match_Manager struct {
	mu          sync.Mutex
	match_queue Queue
	Matches     []*Match
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
		Matches:     make([]*Match, 0),
		match_queue: make(Queue, 0),
	}
}

func (m *Match_Manager) CreateMatch(player1 *Player.Player, player2 *Player.Player, TYpe MatchType, state MatchState) *Match {
	m.mu.Lock()
	defer m.mu.Unlock()

	match := New_match(generateID(), player1, player2, TYpe, state)
	m.Matches = append(m.Matches, match)

	return match

}

func (m *Match_Manager) Start(matchID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, a := range m.Matches {
		if a.ID == matchID {
			a.State = RUNNING
		}
	}
}

func (m *Match_Manager) Finish(matchID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, a := range m.Matches {
		if a.ID == matchID {
			a.State = FINISHED
		}
	}
}

func (m *Match) NextTurn() {
	m.Turn = 3 - m.Turn
}

func (m *Match_Manager) Match_Making() {
	for {
		if len(m.match_queue) >= 2 {
			player1, _ := m.Dequeue()
			player2, _ := m.Dequeue()

			match := m.CreateMatch(&player1, &player2, NORMAL, WAITING)
			m.Start(match.ID)
		}
	}
}

func (m *Match_Manager) Enqueue(val Player.Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.match_queue = append(m.match_queue, val)
}

func (m *Match_Manager) Dequeue() (Player.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.match_queue) == 0 {
		return Player.Player{}, errors.New("empty queue")
	}

	val := (m.match_queue)[0]
	m.match_queue = (m.match_queue)[1:]
	return val, nil
}

/*
- MATCH MAKING POR FILA
- JOGADORES ENTRAM NA FILA E MARCAR COMO 'DISPONIVEL_PARA_JOGO'
- MATCH_MANAGER ESPERA 2 jogadores prontos para criar partida
- Considerar jogos Normal game e ranked game
*/
