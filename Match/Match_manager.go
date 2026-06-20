package match

import (
	"WitchCraft/Cards"
	"WitchCraft/Player"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type Queue []*Player.Player

type Match_Manager struct {
	mu          sync.Mutex
	cond        *sync.Cond // sinaliza quando um jogador entra na fila
	match_queue Queue
	Matches     []*Match
}

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type payload struct {
	Info string `json:"info"`
	Turn int    `json:"turn"`
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
	mm := &Match_Manager{
		Matches:     make([]*Match, 0),
		match_queue: make(Queue, 0),
	}
	mm.cond = sync.NewCond(&mm.mu)
	return mm
}

func (m *Match_Manager) CreateMatch(player1 *Player.Player, player2 *Player.Player, TYpe MatchType, state MatchState, turn int) *Match {
	m.mu.Lock()
	defer m.mu.Unlock()

	match := New_match(generateID(), player1, player2, TYpe, state, turn)
	m.Matches = append(m.Matches, match)

	return match
}

func (m *Match_Manager) RemoveMatch(matchID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, match := range m.Matches {
		if match.ID == matchID {
			m.Matches = append(m.Matches[:i], m.Matches[i+1:]...)
			break
		}
	}
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
		}
	}
}

func (m *Match_Manager) NextTurn(match *Match) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if match.Turn == match.Player1.ID {
		match.Turn = match.Player2.ID
	} else {
		match.Turn = match.Player1.ID
	}
}

func (m *Match_Manager) RemoveFromQueue(playerID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.match_queue {
		if p.ID == playerID {
			m.match_queue = append(m.match_queue[:i], m.match_queue[i+1:]...)
			fmt.Printf("❌ Jogador %d removido da fila\n", playerID)
			return
		}
	}
}

func (m *Match_Manager) Match_Making() {
	for {
		m.mu.Lock()
		// Bloqueia sem consumir CPU até que haja >= 2 jogadores na fila
		for len(m.match_queue) < 2 {
			m.cond.Wait()
		}
		player1 := m.match_queue[0]
		player2 := m.match_queue[1]
		m.match_queue = m.match_queue[2:]
		m.mu.Unlock()

		player1.In_game = true
		player2.In_game = true

		fmt.Println(player1.Conn.LocalAddr())
		fmt.Println(player2.Conn.LocalAddr())

		match := m.CreateMatch(player1, player2, NORMAL, WAITING, player1.ID)
		m.Start(match.ID)
		fmt.Println("The game Start! gameID:", match.ID)
		go m.Run_Game(match)
	}
}

// Enqueue adiciona um jogador à fila e sinaliza o matchmaker.
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
	fmt.Println("Empilhando jogador", val.UserName)
	m.cond.Signal() // acorda o Match_Making se estiver esperando
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
	return val, nil
}

func (m *Match_Manager) Run_Game(match *Match) {
	Data1 := generatePayload(match.Player2.UserName, match.Turn)
	Data2 := generatePayload(match.Player1.UserName, match.Turn)
	Data1_json, _ := json.Marshal(Data1)
	Data2_json, _ := json.Marshal(Data2)

	alert1 := Message{Action: "Game_start", Data: Data1_json}
	alert2 := Message{Action: "Game_start", Data: Data2_json}

	_ = match.Player1.Send(alert1)
	_ = match.Player2.Send(alert2)

	var player1_play, player2_play bool
	var player1_points, player2_points int

	for match.State == RUNNING {
		// Bloqueante: aguarda a próxima ação sem consumir CPU
		msg := <-match.MatchChan
		m.processAction(match, msg, &player1_play, &player2_play)

		if player1_play && player2_play {
			player1_play = false
			player2_play = false

			m.processBattle(match, &player1_points, &player2_points)

			match.Round++
			match.PlayedCard1 = nil
			match.PlayedCard2 = nil

			fmt.Printf("Round: %d\n", match.Round)

			if match.Round >= 3 {
				match.State = FINISHED

				var Winner *Player.Player
				if player1_points > player2_points {
					Winner = match.Player1
				} else {
					Winner = match.Player2
				}

				m.RemoveMatch(match.ID)

				finalMsg := fmt.Sprintf("🛑 Partida finalizada. Vencedor: %s", Winner.UserName)
				finalJSON, _ := json.Marshal(finalMsg)

				match.Player1.In_game = false
				match.Player2.In_game = false

				gameFinish := Message{Action: "game_finish", Data: finalJSON}
				_ = match.Player1.Send(gameFinish)
				_ = match.Player2.Send(gameFinish)

				fmt.Println("FINALIZANDO PARTIDA !!! id da partida:", match.ID)
				return
			}
		}
	}
}

func (m *Match_Manager) processAction(match *Match, msg Match_Message, p1p *bool, p2p *bool) {
	switch msg.Action {
	case "play_card":
		if msg.PlayerId != match.Turn {
			notYourTurn := generatePayload("❌ Não é seu turno.", match.Turn)
			nytb, _ := json.Marshal(notYourTurn)
			target := match.Player1
			if match.Player2.ID == msg.PlayerId {
				target = match.Player2
			}
			_ = target.Send(Message{Action: "game_response", Data: nytb})
			return
		}

		if msg.PlayerId == match.Player1.ID && *p1p {
			already := generatePayload("❌ Você já jogou nesta rodada.", match.Turn)
			ab, _ := json.Marshal(already)
			_ = match.Player1.Send(Message{Action: "game_response", Data: ab})
			return
		}
		if msg.PlayerId == match.Player2.ID && *p2p {
			already := generatePayload("❌ Você já jogou nesta rodada.", match.Turn)
			ab, _ := json.Marshal(already)
			_ = match.Player2.Send(Message{Action: "game_response", Data: ab})
			return
		}

		var play struct {
			Card     *Cards.Card `json:"card"`
			Atribute string      `json:"atribute"`
		}

		if err := json.Unmarshal(msg.Data, &play); err != nil {
			fmt.Println("Erro ao decodificar carta jogada:", err)
			return
		}

		card := play.Card
		atribute := play.Atribute

		if msg.PlayerId == match.Player1.ID {
			match.PlayedCard1 = &PlayedCard{Card: *card, Atribute: atribute}
			*p1p = true
			m.NextTurn(match)
			m.sendToPlayer(match.Player1, match.Turn)
			m.sendToOpponent(msg, match.Player2, match.Turn)
		} else {
			match.PlayedCard2 = &PlayedCard{Card: *card, Atribute: atribute}
			*p2p = true
			m.NextTurn(match)
			m.sendToPlayer(match.Player2, match.Turn)
			m.sendToOpponent(msg, match.Player1, match.Turn)
		}
	}
}

func (m *Match_Manager) processBattle(match *Match, player1_points *int, player2_points *int) {
	card1 := match.PlayedCard1
	card2 := match.PlayedCard2

	if card1 == nil || card2 == nil {
		fmt.Println("Erro: uma das cartas está nula")
		return
	}

	var val1, val2 int
	switch card1.Atribute {
	case "Poder":
		val1 = card1.Card.Power
		val2 = card2.Card.Power
	case "Vida":
		val1 = card1.Card.Life
		val2 = card2.Card.Life
	case "Inteligência":
		val1 = card1.Card.Inteligence
		val2 = card2.Card.Inteligence
	default:
		fmt.Println("Atributo inválido")
		return
	}

	var result string
	if val1 > val2 {
		result = fmt.Sprintf("\n%s venceu a rodada com %s!\n", match.Player1.UserName, card1.Atribute)
		*player1_points++
		match.Turn = match.Player1.ID
	} else if val2 > val1 {
		result = fmt.Sprintf("\n%s venceu a rodada com %s!\n", match.Player2.UserName, card2.Atribute)
		*player2_points++
		match.Turn = match.Player2.ID
	} else {
		result = fmt.Sprintf("\nEmpate na rodada com %s!\n", card1.Atribute)
		*player1_points++
		*player2_points++
	}

	data := generatePayload(result, match.Turn)
	data_json, _ := json.Marshal(data)

	battleResult := Message{Action: "game_response", Data: data_json}
	_ = match.Player1.Send(battleResult)
	_ = match.Player2.Send(battleResult)
}

func (m *Match_Manager) FindMatchByPlayerID(playerId int) *Match {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.Matches {
		match := m.Matches[i]
		if match.Player1.ID == playerId || match.Player2.ID == playerId {
			return match
		}
	}
	return nil
}

func (m *Match_Manager) sendToOpponent(msg Match_Message, opponent *Player.Player, turn int) {
	type PlayedCard struct {
		Card struct {
			Name        string     `json:"name"`
			Power       int        `json:"power"`
			Life        int        `json:"life"`
			Inteligence int        `json:"inteligence"`
			Rarity      Cards.Rare `json:"rarity"`
		} `json:"card"`
		Atribute string `json:"atribute"`
	}

	var played PlayedCard
	if err := json.Unmarshal(msg.Data, &played); err != nil {
		fmt.Println("❌ Erro ao decodificar jogada:", err)
		return
	}

	card := played.Card
	attr := played.Atribute

	match := m.FindMatchByPlayerID(msg.PlayerId)
	var playerName string
	if match != nil {
		if match.Player1.ID == msg.PlayerId {
			playerName = match.Player1.UserName
		} else if match.Player2.ID == msg.PlayerId {
			playerName = match.Player2.UserName
		}
	}

	info := fmt.Sprintf(
		"\n🃏 %s jogou a carta: %s (Power: %d | Life: %d | Inteligência: %d | Raridade: %s)\n🔰 Atributo escolhido: %s\n",
		playerName, card.Name, card.Power, card.Life, card.Inteligence, card.Rarity, attr,
	)

	data := generatePayload(info, turn)
	data_json, _ := json.Marshal(data)

	_ = opponent.Send(Message{Action: "game_response", Data: data_json})
}

func (m *Match_Manager) sendToPlayer(player *Player.Player, turn int) {
	info := "\n✅ Aguarde o oponente.\n"
	data := generatePayload(info, turn)
	data_json, _ := json.Marshal(data)
	_ = player.Send(Message{Action: "game_response", Data: data_json})
}

func generatePayload(info string, turn int) payload {
	return payload{Info: info, Turn: turn}
}
