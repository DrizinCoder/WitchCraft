package api

import (
	"WitchCraft/Cards"
	match "WitchCraft/Match"
	"WitchCraft/Player"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type PlayerResponse struct {
	ID       int    `json:"id"`
	UserName string `json:"username"`
	Login    string `json:"login"`
}

var playerManager = Player.NewManager()
var stock = Cards.NewStock()
var matchManager = match.NewMatchManager()
var logged_players map[string]*Player.Player
var connToPlayer map[net.Conn]*Player.Player
var logged_players_mutex sync.Mutex

func Setup() {

	logged_players = make(map[string]*Player.Player)
	connToPlayer = make(map[net.Conn]*Player.Player)

	stockSize := 1000000
	rand.Seed(time.Now().UnixNano())

	rarities := []Cards.Rare{Cards.BRONZE, Cards.SILVER, Cards.GOLD, Cards.DIAMOND}
	names := []string{
		"Fireball", "Icebolt", "Goblin", "Dragon", "Knight", "Elf",
		"Iceball", "White_Knight", "Giant_goblin", "Dragon_Black", "Elder_witch", "Elf_elder",
	}

	for i := 0; i < stockSize; i++ {
		base := names[i%len(names)]
		name := fmt.Sprintf("%s_%d", base, i)

		power := rand.Intn(20) + 1 // 1..20
		life := rand.Intn(25) + 1  // 1..25
		intel := rand.Intn(15) + 1 // 1..15
		rarity := rarities[rand.Intn(len(rarities))]

		stock.CreateCard(name, power, life, rarity, intel)
	}

	go matchManager.Match_Making()

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer listener.Close()

	for {
		fmt.Println(len(stock.Deck))
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao conectar:", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg Message
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("Erro ao decodificar mensagem:", err)
			handleDisconnect(conn)
			return
		}

		switch msg.Action {
		case "create_player":
			createPlayerHandler(msg, encoder)
		case "login_player":
			loginPlayerHandler(msg, encoder, conn)
		case "open_pack":
			openPackHandler(msg, encoder)
		case "search_player":
			getPlayerHandler(msg, encoder)
		case "enqueue_player":
			enqueue(msg, encoder)
		case "see_inventory":
			getInventoryHandler(msg, encoder)
		case "Game_Action":
			gameAction(msg)
		case "set_deck":
			setDeckHandler(msg, encoder)
		case "get_deck":
			getDeckHandler(msg, encoder)
		}
	}
}

func createPlayerHandler(msg Message, encoder *json.Encoder) {

	type req struct {
		Username string `json:"username"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	player, err := playerManager.Create_Player(r.Username, r.Login, r.Password)

	if err != nil {
		send_error(err, encoder)
		return
	}

	response := PlayerResponse{
		ID:       player.ID,
		UserName: player.UserName,
		Login:    player.Login,
	}

	response_json, _ := json.Marshal(response)

	final_msg := Message{
		Action: "create_player_response",
		Data:   response_json,
	}

	encoder.Encode(final_msg)
}

func loginPlayerHandler(msg Message, encoder *json.Encoder, conn net.Conn) {

	type req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	logged_players_mutex.Lock()
	_, exists := logged_players[r.Login] // Aqui pode ter concorrência
	logged_players_mutex.Unlock()

	if exists {
		err := errors.New("user already logged")
		send_error(err, encoder)
		return
	}

	player, err := playerManager.Login(r.Login, r.Password, conn)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	logged_players_mutex.Lock()
	logged_players[r.Login] = player
	connToPlayer[conn] = player
	logged_players_mutex.Unlock()

	response := PlayerResponse{
		ID:       player.ID,
		UserName: player.UserName,
		Login:    player.Login,
	}

	response_json, _ := json.Marshal(response)

	final_msg := Message{
		Action: "login_player_response",
		Data:   response_json,
	}

	encoder.Encode(final_msg)
}

func openPackHandler(msg Message, encoder *json.Encoder) {

	type req struct {
		PlayerID int `json:"id"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	pack, err := playerManager.Open_pack(r.PlayerID, stock)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	pack_json, _ := json.Marshal(pack)

	final_msg := Message{
		Action: "open_pack_response",
		Data:   pack_json,
	}

	encoder.Encode(final_msg)
}

func getPlayerHandler(msg Message, encoder *json.Encoder) {

	type req struct {
		PlayerID int `json:"id"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	player, err := playerManager.Search_Player_ByID(r.PlayerID)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	response := PlayerResponse{
		ID:       player.ID,
		UserName: player.UserName,
		Login:    player.Login,
	}

	response_json, _ := json.Marshal(response)

	final_msg := Message{
		Action: "search_player_response",
		Data:   response_json,
	}

	encoder.Encode(final_msg)
}

func setDeckHandler(msg Message, encoder *json.Encoder) {
	// Estrutura esperada do cliente
	type req struct {
		PlayerID int           `json:"player_id"`
		Deck     []*Cards.Card `json:"deck"`
	}

	var r req
	err := json.Unmarshal(msg.Data, &r)
	if err != nil {
		send_error(err, encoder)
		return
	}

	// Busca o jogador
	player, err := playerManager.Search_Player_ByID(r.PlayerID)
	if err != nil {
		send_error(err, encoder)
		return
	}

	// Valida se todas as cartas do deck estão no inventário do jogador
	for _, deckCard := range r.Deck {
		found := false
		for _, invCard := range player.Cards {
			if deckCard.Name == invCard.Name &&
				deckCard.Power == invCard.Power &&
				deckCard.Life == invCard.Life &&
				deckCard.Inteligence == invCard.Inteligence &&
				deckCard.Rarity == invCard.Rarity {
				found = true
				break
			}
		}
		if !found {
			send_error(errors.New("uma ou mais cartas do deck não estão no inventário"), encoder)
			return
		}
	}

	// Atualiza o deck do jogador
	player.GameDeck = r.Deck

	// Resposta de sucesso
	payload := map[string]string{"success": "deck definido com sucesso"}
	data, _ := json.Marshal(payload)
	final_msg := Message{
		Action: "set_deck_response",
		Data:   data,
	}
	encoder.Encode(final_msg)
}

func enqueue(msg Message, encoder *json.Encoder) {

	type req struct {
		PlayerID int `json:"id"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	player, err := playerManager.Search_Player_ByID(r.PlayerID)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	err = matchManager.Enqueue(player)

	if err != nil {
		send_error(err, encoder)
		return
	}

	payload := map[string]string{"Player enqueued": player.UserName}

	data, _ := json.Marshal(payload)

	final_msg := Message{
		Action: "enqueue_response",
		Data:   json.RawMessage(data),
	}

	encoder.Encode(final_msg)
}

func gameAction(msg Message) {
	type req struct {
		PlayerID int             `json:"PlayerID"`
		Action   string          `json:"action"`
		Payload  json.RawMessage `json:"data"`
	}
	var r req
	json.Unmarshal(msg.Data, &r)

	fmt.Println("Recebi Game_Action:", string(msg.Data))

	match1 := matchManager.FindMatchByPlayerID(r.PlayerID)
	fmt.Println(match1)
	fmt.Println("Procurando match para player:", r.PlayerID)
	if match1 != nil {
		match1.MatchChan <- match.Match_Message{
			PlayerId: r.PlayerID,
			Action:   r.Action,
			Data:     r.Payload,
		}
	}
}

func getInventoryHandler(msg Message, encoder *json.Encoder) {
	type req struct {
		PlayerID int `json:"id"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	Cards, _ := playerManager.Get_inventory(r.PlayerID)

	cards_json, _ := json.Marshal(Cards)

	final_msg := Message{
		Action: "see_inventory_response",
		Data:   cards_json,
	}

	encoder.Encode(final_msg)
}

func getDeckHandler(msg Message, encoder *json.Encoder) {
	type req struct {
		PlayerID int `json:"id"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	Cards, _ := playerManager.Get_deck(r.PlayerID)

	cards_json, _ := json.Marshal(Cards)

	final_msg := Message{
		Action: "get_deck_response",
		Data:   cards_json,
	}

	encoder.Encode(final_msg)
}

func handleDisconnect(conn net.Conn) {
	logged_players_mutex.Lock()
	player, ok := connToPlayer[conn]
	logged_players_mutex.Unlock()

	if !ok {
		return
	}

	fmt.Printf("⚠️ Jogador %s desconectou.\n", player.UserName)

	matchManager.RemoveFromQueue(player.ID)

	if player.In_game {
		match := matchManager.FindMatchByPlayerID(player.ID)
		if match != nil {
			var opponent *Player.Player
			if match.Player1.ID == player.ID {
				opponent = match.Player2
			} else {
				opponent = match.Player1
			}

			player.In_game = false
			opponent.In_game = false

			finalPayload := fmt.Sprintf("🛑 Partida finalizada. O jogador %s desconectou.", player.UserName)
			finalPayloadJSON, _ := json.Marshal(finalPayload)

			alert := Message{
				Action: "game_finish",
				Data:   finalPayloadJSON,
			}

			if opponent.Conn != nil {
				enc := json.NewEncoder(opponent.Conn)
				_ = enc.Encode(alert)
			}

			matchManager.RemoveMatch(match.ID)
		}
	}

	logged_players_mutex.Lock()
	delete(logged_players, player.UserName)
	delete(connToPlayer, conn)
	logged_players_mutex.Unlock()
}

func send_error(err error, encoder *json.Encoder) {
	payload := map[string]string{"error": err.Error()}

	data, _ := json.Marshal(payload)

	erro_msg := Message{
		Action: "error_response",
		Data:   json.RawMessage(data),
	}

	encoder.Encode(erro_msg)
}
