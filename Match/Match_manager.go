package match

import (
	"WitchCraft/Player"
	"encoding/json"
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

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
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

func (m *Match_Manager) NextTurn(match *Match) {
	match.Turn = 3 - match.Turn
}

func (m *Match_Manager) Match_Making() {
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
	encoder1 := json.NewEncoder(match.Player1.Conn)
	encoder2 := json.NewEncoder(match.Player2.Conn)

	Data1, _ := json.Marshal(match.Player2.UserName)
	Data2, _ := json.Marshal(match.Player1.UserName)

	alert1 := Message{
		Action: "Game_start",
		Data:   Data1,
	}

	alert2 := Message{
		Action: "Game_start",
		Data:   Data2,
	}

	encoder1.Encode(alert1)
	encoder2.Encode(alert2)

	for match.State == RUNNING {
		select {
		case msg := <-match.MatchChan:
			m.processAction(match, msg, encoder1, encoder2)
		}
	}

}

func (m *Match_Manager) processAction(match *Match, msg Match_Message, encoder1 *json.Encoder, encoder2 *json.Encoder) {
	switch msg.Action {
	case "play_card":
		fmt.Println("Jogador", msg.PlayerId, "jogou carta:", msg.Data)
		m.NextTurn(match)
		if match.Turn == 1 {
			m.sendToOpponent(msg, encoder2)
		} else {
			m.sendToOpponent(msg, encoder1)
		}
	case "end_turn":
		match.Turn = 3 - match.Turn
		// m.sendToOpponent(match, msg.PlayerId, msg)
	}
}

func (m *Match_Manager) FindMatchByPlayerID(playerId int) *Match {
	for i := range m.Matches {
		match := m.Matches[i]
		if match.Player1.ID == playerId || match.Player2.ID == playerId {
			return match
		}
	}
	return nil
}

func (m *Match_Manager) sendToOpponent(msg Match_Message, encoder *json.Encoder) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := fmt.Sprintf("Jogador %d jogou carta: %s", msg.PlayerId, msg.Data)

	data_json, _ := json.Marshal(data)

	payload := Message{
		Action: "game_response",
		Data:   data_json,
	}

	encoder.Encode(payload)
}
