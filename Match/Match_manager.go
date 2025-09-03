package match

import (
	"WitchCraft/Cards"
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
	return &Match_Manager{
		Matches:     make([]*Match, 0),
		match_queue: make(Queue, 0),
	}
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
			//Lembrar que ainda tem que tirar a partida finalizada da lista de matches ativos
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
			match := m.CreateMatch(player1, player2, NORMAL, WAITING, player1.ID)
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

	Data1 := generatePayload(match.Player2.UserName, match.Turn)
	Data2 := generatePayload(match.Player1.UserName, match.Turn)
	Data1_json, _ := json.Marshal(Data1)
	Data2_json, _ := json.Marshal(Data2)

	alert1 := Message{
		Action: "Game_start",
		Data:   Data1_json,
	}

	alert2 := Message{
		Action: "Game_start",
		Data:   Data2_json,
	}

	encoder1.Encode(alert1)
	encoder2.Encode(alert2)

	player1_play := false
	player2_play := false
	player1_points := 0
	player2_points := 0

	for match.State == RUNNING {
		select {
		case msg := <-match.MatchChan:
			m.processAction(match, msg, encoder1, encoder2, &player1_play, &player2_play)
		default:
			if player1_play && player2_play {
				player1_play = false
				player2_play = false

				m.processBattle(match, encoder1, encoder2, &player1_points, &player2_points)

				match.Round++
				match.PlayedCard1 = nil
				match.PlayedCard2 = nil

				fmt.Printf("Round: %d\n", match.Round)

				if match.Round >= 3 {
					m.Finish(match.ID)
					if match.Round >= 3 {
						match.State = FINISHED

						var Winner *Player.Player
						if player1_points > player2_points {
							Winner = match.Player1
						} else {
							Winner = match.Player2
						}

						m.RemoveMatch(match.ID)

						finalPayload1 := fmt.Sprintf("🛑 Partida finalizada. Vencedor: %s", Winner.UserName)
						finalPayload2 := fmt.Sprintf("🛑 Partida finalizada. Vencedor: %s", Winner.UserName)

						match.Player1.In_game = false
						match.Player2.In_game = false

						finalPayloadJSON1, _ := json.Marshal(finalPayload1)
						finalPayloadJSON2, _ := json.Marshal(finalPayload2)

						alert1 := Message{
							Action: "game_finish",
							Data:   finalPayloadJSON1,
						}
						alert2 := Message{
							Action: "game_finish",
							Data:   finalPayloadJSON2,
						}

						encoder1.Encode(alert1)
						encoder2.Encode(alert2)

						fmt.Println("FINALIZANDO PARTIDA !!!")
						return
					}
					return
				}
			}
		}
	}

}

func (m *Match_Manager) processAction(match *Match, msg Match_Message, encoder1 *json.Encoder, encoder2 *json.Encoder, p1p *bool, p2p *bool) {
	switch msg.Action {
	case "play_card":
		// 1) Checar se é o turno do jogador
		if msg.PlayerId != match.Turn {
			// Opcional: responder só para quem tentou jogar fora do turno
			notYourTurn := generatePayload("❌ Não é seu turno.", match.Turn)
			nytb, _ := json.Marshal(notYourTurn)
			target := encoder1
			if match.Player2.ID == msg.PlayerId {
				target = encoder2
			}
			_ = target.Encode(Message{Action: "game_response", Data: nytb})
			return
		}

		// 2) Bloquear segunda jogada na mesma rodada
		if msg.PlayerId == match.Player1.ID && *p1p {
			already := generatePayload("❌ Você já jogou nesta rodada.", match.Turn)
			ab, _ := json.Marshal(already)
			_ = encoder1.Encode(Message{Action: "game_response", Data: ab})
			return
		}
		if msg.PlayerId == match.Player2.ID && *p2p {
			already := generatePayload("❌ Você já jogou nesta rodada.", match.Turn)
			ab, _ := json.Marshal(already)
			_ = encoder2.Encode(Message{Action: "game_response", Data: ab})
			return
		}

		var play struct {
			Card     *Cards.Card `json:"card"`
			Atribute string      `json:"atribute"`
		}

		err := json.Unmarshal(msg.Data, &play)
		if err != nil {
			fmt.Println("Erro ao decodificar carta jogada:", err)
			return
		}

		card := play.Card
		atribute := play.Atribute

		if msg.PlayerId == match.Player1.ID {
			// player1 jogou
			match.PlayedCard1 = &PlayedCard{Card: *card, Atribute: atribute}
			*p1p = true

			m.NextTurn(match)
			m.sendToPlayer(encoder1, match.Turn)        // envia "aguarde" para o player1
			m.sendToOpponent(msg, encoder2, match.Turn) // envia carta para player2
		} else {
			// player2 jogou
			match.PlayedCard2 = &PlayedCard{Card: *card, Atribute: atribute}
			*p2p = true

			m.NextTurn(match)
			m.sendToPlayer(encoder2, match.Turn)        // envia "aguarde" para player2
			m.sendToOpponent(msg, encoder1, match.Turn) // envia carta para player1
		}
	}
}

func (m *Match_Manager) processBattle(match *Match, encoder1 *json.Encoder, encoder2 *json.Encoder, player1_points *int, player2_points *int) {
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

	payload := Message{
		Action: "game_response",
		Data:   data_json,
	}

	encoder1.Encode(payload)
	encoder2.Encode(payload)
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

func (m *Match_Manager) sendToOpponent(msg Match_Message, encoder *json.Encoder, turn int) {
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
	err := json.Unmarshal(msg.Data, &played)
	if err != nil {
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

	initial := fmt.Sprintf("\n🃏 %s jogou a carta: %s (Power: %d | Life: %d | Inteligência: %d | Raridade: %s)\n🔰 Atributo escolhido: %s\n",
		playerName, card.Name, card.Power, card.Life, card.Inteligence, card.Rarity, attr)

	data := generatePayload(initial, turn)

	data_json, _ := json.Marshal(data)

	payload := Message{
		Action: "game_response",
		Data:   data_json,
	}

	_ = encoder.Encode(payload)
}

func (m *Match_Manager) sendToPlayer(encoder *json.Encoder, turn int) {
	initial := "\n✅ Aguarde o oponente.\n"
	data := generatePayload(initial, turn)

	data_json, _ := json.Marshal(data)

	payload := Message{
		Action: "game_response",
		Data:   data_json,
	}

	_ = encoder.Encode(payload)
}

func generatePayload(info string, turn int) payload {
	Data := payload{
		Info: info,
		Turn: turn,
	}

	return Data
}
