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

func (m *Match) Start() {
	m.State = RUNNING
}

func (m *Match) Finish() {
	m.State = FINISHED
}

func (m *Match) NextTurn() {
	m.Turn = 3 - m.Turn
}

func (q *Queue) Enqueue(val Player.Player) {
	*q = append(*q, val)
}

func (q *Queue) Dequeue() (Player.Player, error) {
	if len(*q) == 0 {
		return Player.Player{}, errors.New("empty queue")
	}

	val := (*q)[0]
	*q = (*q)[1:]
	return val, nil
}

/*
- MATCH MAKING POR FILA
- JOGADORES ENTRAM NA FILA E MARCAR COMO 'DISPONIVEL_PARA_JOGO'
- MATCH_MANAGER ESPERA 2 jogadores prontos para criar partida
- Considerar jogos Normal game e ranked game
*/
