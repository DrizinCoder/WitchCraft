package api

import (
	"WitchCraft/Cards"
	match "WitchCraft/Match"
	"WitchCraft/Player"
	"encoding/json"
	"errors"
	"fmt"
	"net"
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

func Setup() {

	logged_players = make(map[string]*Player.Player)

	stock.CreateCard("Fireball", 10, 5, Cards.GOLD)
	stock.CreateCard("Icebolt", 8, 6, Cards.SILVER)
	stock.CreateCard("Goblin", 5, 10, Cards.BRONZE)
	stock.CreateCard("Dragon", 20, 20, Cards.DIAMOND)
	stock.CreateCard("Knight", 12, 15, Cards.SILVER)
	stock.CreateCard("Elf", 7, 8, Cards.BRONZE)

	stock.CreateCard("Iceball", 10, 5, Cards.GOLD)
	stock.CreateCard("White_Knight", 8, 6, Cards.SILVER)
	stock.CreateCard("Giant_goblin", 5, 10, Cards.BRONZE)
	stock.CreateCard("Dragon_Black", 20, 20, Cards.DIAMOND)
	stock.CreateCard("Elder_witch", 12, 15, Cards.SILVER)
	stock.CreateCard("Elf_elder", 7, 8, Cards.BRONZE)

	go matchManager.Match_Making()

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer listener.Close()

	for {
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
			return // this can be better of course!
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
		case "ping":
			pingHandler(encoder)
		case "Game_Action":
			gameAction(msg)
		case "set_deck":
			setDeckHandler(msg, encoder)
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

	_, exists := logged_players[r.Login]

	if exists {
		err := errors.New("user already logged")
		send_error(err, encoder)
		return
	}

	player, err := playerManager.Login(r.Login, r.Password, conn)
	fmt.Println(player.Conn.LocalAddr())

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
		return
	}

	logged_players[r.Login] = player

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

func pingHandler(encoder *json.Encoder) {
	a := "a"
	data_json, _ := json.Marshal(a)
	final_msg := Message{
		Action: "pong_response",
		Data:   data_json,
	}

	encoder.Encode(final_msg)
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
