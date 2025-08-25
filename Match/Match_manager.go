package match

import (
	"WitchCraft/Player"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Queue []*Player.Player

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

	for i := range m.Matches {
		if m.Matches[i].ID == matchID {
			m.Matches[i].State = RUNNING
		}
	}
}

func (m *Match_Manager) Finish(matchID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Matches {
		if m.Matches[i].ID == matchID {
			m.Matches[i].State = FINISHED
			//Lembrar que ainda tem que tirar a partida finalizada da lista de matches ativos
		}
	}
}

func (m *Match_Manager) NextTurn(matchID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Matches {
		if m.Matches[i].ID == matchID {
			m.Matches[i].Turn = 3 - m.Matches[i].Turn
		}
	}
}

func (m *Match_Manager) Match_Making() { // Retornar a referencia do match criado e passar para uma goroutine que tomará conta do game
	for {
		if len(m.match_queue) >= 2 {
			player1, err1 := m.Dequeue()
			player2, err2 := m.Dequeue()
			if err1 != nil || err2 != nil {
				continue
			}

			player1.In_game = true
			player2.In_game = true
			fmt.Println(player1.Conn.LocalAddr())
			fmt.Println(player2.Conn.LocalAddr())
			match := m.CreateMatch(player1, player2, NORMAL, WAITING)
			m.Start(match.ID)
			println("The game Start!")
			go m.Run_Game(match) //  Go routine que tomará conta do jogo
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (m *Match_Manager) Enqueue(val *Player.Player) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if val.In_game {
		return errors.New("you can't join in the queue, alredy in a current game")
	}

	for _, p := range m.match_queue {
		if p.ID == val.ID {
			return errors.New("player alredy in match queue")
		}
	}

	m.match_queue = append(m.match_queue, val)
	println("Empilhando jogador" + val.UserName)
	return nil
}

func (m *Match_Manager) Dequeue() (*Player.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.match_queue) == 0 {
		return &Player.Player{}, errors.New("empty queue")
	}

	val := (m.match_queue)[0]
	m.match_queue = (m.match_queue)[1:]
	println("removendo jogador")
	return val, nil
}

func (m *Match_Manager) Run_Game(match *Match) {

	for match.State == RUNNING {
		select {
		case msg := <-match.MatchChan:
			m.processAction(match, msg)
		}
	}

}

func (m *Match_Manager) processAction(match *Match, msg Match_Message) {
	switch msg.Action {
	case "play_card":
		fmt.Println("Jogador", msg.PlayerId, "jogou carta:", msg.Data)
		m.sendToOpponent(match, msg.PlayerId, msg)
	case "end_turn":
		match.Turn = 3 - match.Turn
		m.sendToOpponent(match, msg.PlayerId, msg)
	}
}

func (m *Match_Manager) FindMatchByPlayerID(matchId int) *Match {

	for i := range m.Matches {
		if m.Matches[i].ID == matchId {
			return m.Matches[i]
		}
	}

	return nil
}

func (m *Match_Manager) sendToOpponent(match *Match, playerID int, msg Match_Message) {
	// Envia ao oponente a resposta
}
