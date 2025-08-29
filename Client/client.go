package client

import (
	"WitchCraft/Cards"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type Game_Message struct {
	PlayerID int             `json:"PlayerID"`
	Action   string          `json:"action"`
	Data     json.RawMessage `json:"data"`
}

type Req_player struct {
	Username string `json:"username"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Req_login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Req_id struct {
	ID int `json:"id"`
}

type payload struct {
	Info string `json:"info"`
	Turn int    `json:"turn"`
}

var session_id int
var start time.Time
var channel chan int
var encoder *json.Encoder

var playerInventory []*Cards.Card
var playerInventoryMutex sync.RWMutex

var playerDeck []*Cards.Card
var playerDeckMutex sync.RWMutex

var gameTurn int
var gameTurnMutex sync.RWMutex

func Setup() {

	serverAddr := os.Getenv("SERVER_ADDR")
	conn, err := net.Dial("tcp", serverAddr)

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder = json.NewEncoder(conn)

	go handleConnection(decoder)

	channel = make(chan int, 1)

	for {
		fmt.Println("\n==============================")
		fmt.Println(" 🎮 WitchCraft - Menu Principal ")
		fmt.Println("==============================")
		fmt.Println("1️⃣  - Registrar Jogador")
		fmt.Println("2️⃣  - Login")
		fmt.Println("3️⃣  - Abrir Pacote de Cartas")
		fmt.Println("4️⃣  - Buscar Jogador")
		fmt.Println("5️⃣  - Entrar na Fila")
		fmt.Println("6️⃣  - Ver inventário")
		fmt.Println("7️⃣  - Medir Ping")
		fmt.Println("0️⃣  - Sair")
		fmt.Println("------------------------------")
		fmt.Print("👉 Escolha a sua próxima ação: ")

		var action int
		go func() {
			fmt.Scanln(&action)
			if action == 99 {
				action = 20
			}
			channel <- action
		}()

		change := <-channel
		switch change {
		case 1:
			createPlayer(encoder)
		case 2:
			loginPlayer(encoder)
		case 3:
			openPack(encoder)
		case 4:
			searchPlayer(encoder)
		case 5:
			enqueue(encoder)
		case 6:
			seeInventory()
		case 7:
			ping(encoder)
		case 0:
			fmt.Println("👋 Saindo do jogo... Até logo!")
			return
		case 99:
			GameMenu(encoder)
		default:
			fmt.Println("❌ Opção inválida, tente novamente.")
		}
	}

}

func handleConnection(decoder *json.Decoder) {

	for {
		var payload Message
		err := decoder.Decode(&payload)
		if err != nil {
			fmt.Println("Erro ao ler mensagem do servidor.")
			return
		}
		switch payload.Action {
		case "create_player_response":
			handleCreatePlayerResponse(payload.Data)
		case "login_player_response":
			handleLoginPlayerResponse(payload.Data)
		case "open_pack_response":
			handleOpenPackResponse(payload.Data)
		case "search_player_response":
			handleSearchPlayerResponse(payload.Data)
		case "enqueue_response":
			handleEnqueueResponse(payload.Data)
		case "error_response":
			handleErrorResponse(payload.Data)
		case "see_inventory_response":
			handleSeeInventoryResponse(payload.Data)
		case "pong_response":
			handlePongResponse()
		case "set_deck_response":
			handleSetDeckResponse(payload.Data)
		case "Game_start":
			handleGameStartResponse(payload.Data)
		case "game_response":
			handleGameResponse(payload.Data)
		case "get_deck_response":
			handleGetDeckResponse(payload.Data)
		}
	}

}

func createPlayer(encoder *json.Encoder) {
	username := prompt("👤 Digite o nome do jogador: ")
	login := prompt("📧 Digite o login: ")
	password := prompt("🔑 Digite a senha: ")

	payload := Req_player{
		Username: username,
		Login:    login,
		Password: password,
	}

	sendRequest(encoder, "create_player", payload)
}

func loginPlayer(encoder *json.Encoder) {
	login := prompt("📧 Digite o login: ")
	password := prompt("🔑 Digite a senha: ")

	payload := Req_login{
		Login:    login,
		Password: password,
	}

	sendRequest(encoder, "login_player", payload)
}

func openPack(encoder *json.Encoder) {
	if session_id == 0 {
		fmt.Println("❌ Opção inválida, você deve estar logado para completar essa ação")
		return
	}

	payload := Req_id{
		ID: session_id,
	}

	sendRequest(encoder, "open_pack", payload)

}

func searchPlayer(encoder *json.Encoder) {
	if session_id == 0 {
		return
	}

	payload := Req_id{
		ID: session_id,
	}

	sendRequest(encoder, "search_player", payload)
}

func enqueue(encoder *json.Encoder) {
	if session_id == 0 {
		fmt.Println("❌ Opção inválida, você deve estar logado para completar essa ação")
		return
	}

	if len(playerDeck) == 0 {
		fmt.Println("❌ Opção inválida, você deve montar seu deck de jogo")
		return
	}

	payload := Req_id{
		ID: session_id,
	}

	sendRequest(encoder, "enqueue_player", payload)
}

func seeInventory() {
	if session_id == 0 {
		fmt.Println("❌ Opção inválida, você deve estar logado para completar essa ação")
		return
	}

	playerInventoryMutex.RLock()
	if len(playerInventory) == 0 {
		playerInventoryMutex.RUnlock()
		fmt.Println("Sem cartas no inventário.")
		return
	}

	fmt.Println("Inventário de cartas:")
	for i, c := range playerInventory {
		fmt.Printf("%d️⃣ - %s (Power: %d, Life: %d, Inteligence: %d, Rarity: %s)\n",
			i+1, c.Name, c.Power, c.Life, c.Inteligence, c.Rarity)
	}

	inventoryCopy := make([]*Cards.Card, len(playerInventory))
	copy(inventoryCopy, playerInventory)
	playerInventoryMutex.RUnlock()

	fmt.Println("\nDeseja montar seu deck de 3 cartas? (s/n)")
	var choice string
	fmt.Scanln(&choice)
	if choice == "s" || choice == "S" {
		chooseDeck(inventoryCopy)
	}
}

func chooseDeck(inventory []*Cards.Card) {
	if len(inventory) < 3 {
		fmt.Println("Você não tem cartas suficientes para montar um deck.")
		return
	}

	fmt.Println("Escolha 3 cartas para seu deck de batalha:")
	for i, c := range inventory {
		fmt.Printf("%d️⃣ - %s (Power: %d, Life: %d, Inteligence: %d, Rarity: %s)\n",
			i+1, c.Name, c.Power, c.Life, c.Inteligence, c.Rarity)
	}

	selectedIndexes := make([]int, 0, 3)
	for len(selectedIndexes) < 3 {
		fmt.Printf("Digite o número da carta %d: ", len(selectedIndexes)+1)
		var choice int
		fmt.Scanln(&choice)

		if choice < 1 || choice > len(inventory) {
			fmt.Println("Escolha inválida, tente novamente.")
			continue
		}

		duplicate := false
		for _, idx := range selectedIndexes {
			if idx == choice-1 {
				duplicate = true
				break
			}
		}

		if duplicate {
			fmt.Println("Você já escolheu essa carta.")
			continue
		}

		selectedIndexes = append(selectedIndexes, choice-1)
	}

	var selectedCards []*Cards.Card
	for _, idx := range selectedIndexes {
		selectedCards = append(selectedCards, inventory[idx])
	}

	playerDeckMutex.Lock()
	playerDeck = selectedCards
	playerDeckMutex.Unlock()

	payload := map[string]interface{}{
		"player_id": session_id,
		"deck":      selectedCards,
	}
	sendRequest(encoder, "set_deck", payload)

	fmt.Println("Deck escolhido com sucesso!")
}

func ping(encoder *json.Encoder) {
	start = time.Now()

	payload := ""
	sendRequest(encoder, "ping", payload)
}

func play_card(encoder *json.Encoder) {
	if len(playerDeck) == 0 {
		fmt.Println("⚠️ Você precisa montar seu deck antes de jogar!")
		return
	}

	fmt.Println("Escolha uma carta do seu deck para jogar:")
	for i, c := range playerDeck {
		fmt.Printf("%d️⃣ - %s (Power: %d, Life: %d, Inteligence: %d, Rarity: %s)\n",
			i+1, c.Name, c.Power, c.Life, c.Inteligence, c.Rarity)
	}

	var choice int
	var choice2 int
	for {
		fmt.Print("Digite o número da carta: ")
		fmt.Scanln(&choice)

		if choice < 1 || choice > len(playerDeck) {
			fmt.Println("Escolha inválida, tente novamente.")
			continue
		}
		break
	}
	for {
		fmt.Print("Digite o atributo que deseja competir")
		fmt.Printf("1 - Inteligencia\n2 - Poder\n3 - Vida")
		fmt.Scanln(&choice2)

		if choice2 < 1 || choice2 > 3 {
			fmt.Println("Escolha inválida, tente novamente.")
			continue
		}
		break
	}

	type playCard struct {
		Card     *Cards.Card `json:"card"`
		Atribute string      `json:"atribute"`
	}

	var req playCard

	req.Card = playerDeck[choice-1]
	switch choice2 {
	case 1:
		req.Atribute = "Inteligência"
	case 2:
		req.Atribute = "Poder"
	case 3:
		req.Atribute = "Vida"
	}

	req_json, _ := json.Marshal(req)

	msg := Game_Message{
		PlayerID: session_id,
		Action:   "play_card",
		Data:     req_json,
	}

	msg_json, _ := json.Marshal(msg)

	payload := Message{
		Action: "Game_Action",
		Data:   msg_json,
	}

	err := encoder.Encode(payload)
	if err != nil {
		fmt.Println("Erro ao enviar ação para o servidor:", err)
	} else {
		fmt.Printf("🃏 Você jogou a carta: %s\n", req.Card)
	}
}

func handleCreatePlayerResponse(data json.RawMessage) {
	type req struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
		Login    string `json:"login"`
	}

	var resp req
	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
		return
	}

	fmt.Println("Jogador criado com sucesso. Retorno do servidor: ")
	fmt.Println("ID:", resp.ID)
	fmt.Println("Username", resp.UserName)
	fmt.Println("Login", resp.Login)

}

func handleLoginPlayerResponse(data json.RawMessage) {
	type req struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
		Login    string `json:"login"`
	}

	var resp req
	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
		return
	}

	fmt.Println("Jogador efetuado com sucesso. Retorno do servidor: ")
	fmt.Println("ID:", resp.ID)
	fmt.Println("Username", resp.UserName)

	session_id = resp.ID
	gameTurn = 0

	go func() {
		payload := Req_id{ID: session_id}
		sendRequest(encoder, "see_inventory", payload)
	}()

	go func() {
		payload := Req_id{ID: session_id}
		sendRequest(encoder, "get_deck", payload)
	}()
}

func handleOpenPackResponse(data json.RawMessage) {

	var cards []Cards.Card
	err := json.Unmarshal(data, &cards)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados: ", err)
		return
	}

	playerInventoryMutex.Lock()
	fmt.Println("Pacote Aberto. Cartas: ")
	for i, c := range cards {
		fmt.Printf("- %s (Power: %d, Life: %d, Rarity: %s)\n",
			c.Name, c.Power, c.Life, c.Rarity)
		playerInventory = append(playerInventory, &cards[i])
	}
	playerInventoryMutex.Unlock()
}

func handleSearchPlayerResponse(data json.RawMessage) {
	type req struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
		Login    string `json:"login"`
	}

	var resp req
	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
		return
	}

	fmt.Println("Jogador encontrado. Retorno do servidor: ")
	fmt.Println("ID:", resp.ID)
	fmt.Println("Username", resp.UserName)
	fmt.Println("Login", resp.Login)
}

func handleEnqueueResponse(data json.RawMessage) {

	var resp map[string]string

	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
	}

	fmt.Println(resp["Player enqueued"])
}

func handleErrorResponse(data json.RawMessage) {

	var resp map[string]string

	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
	}

	fmt.Println(resp["error"])
}

func handleSeeInventoryResponse(data json.RawMessage) {
	var cards []Cards.Card
	err := json.Unmarshal(data, &cards)
	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados: ", err)
		return
	}

	playerInventoryMutex.Lock()
	playerInventory = make([]*Cards.Card, len(cards))
	for i := range cards {
		playerInventory[i] = &cards[i]
	}
	playerInventoryMutex.Unlock()

	fmt.Println("Inventário carregado com sucesso! Total de cartas:", len(playerInventory))
}

func handleGetDeckResponse(data json.RawMessage) {
	var cards []Cards.Card
	err := json.Unmarshal(data, &cards)
	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados: ", err)
		return
	}

	playerDeck = make([]*Cards.Card, len(cards))
	for i := range cards {
		playerDeck[i] = &cards[i]
	}
	fmt.Println("Deck de jogo carregado com sucesso!")
}

func handlePongResponse() {
	elapsed := time.Since(start)
	fmt.Printf("Ping: %s\n", elapsed)
}

func handleSetDeckResponse(data json.RawMessage) {
	var resp map[string]string
	err := json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("Erro ao decodificar resposta do servidor:", err)
		return
	}

	if msg, ok := resp["success"]; ok {
		fmt.Println("✅", msg)
	} else {
		fmt.Println("❌ Algo deu errado ao definir o deck.")
	}
}

func handleGameResponse(data json.RawMessage) {

	var resp payload

	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados: ", err)
		return
	}

	gameTurnMutex.Lock()
	gameTurn = resp.Turn
	gameTurnMutex.Unlock()

	if resp.Turn == session_id {
		fmt.Printf("%s. Seu turno, realize sua jogada", resp.Info)
	} else {
		fmt.Printf("%s. Aguarde seu turno para jogar", resp.Info)
	}

}

func handleGameStartResponse(data json.RawMessage) {
	var resp payload

	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados: ", err)
		return
	}

	gameTurnMutex.Lock()
	gameTurn = resp.Turn
	gameTurnMutex.Unlock()

	channel <- 99
	fmt.Printf("O Jogo iniciou! Pareado com o jogador: %s", resp.Info)

	if resp.Turn == session_id {
		fmt.Printf("%s. Seu turno, realize sua jogada", resp.Info)
	} else {
		fmt.Printf("%s. Aguarde seu turno para jogar", resp.Info)
	}

}

func sendRequest(encoder *json.Encoder, action string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Erro ao serializar payload:", err)
		return
	}

	req := Message{
		Action: action,
		Data:   data,
	}

	err = encoder.Encode(req)
	if err != nil {
		fmt.Println("Erro ao enviar requisição:", err)
		return
	}
}

func GameMenu(encoder *json.Encoder) {
	for {
		fmt.Println("\n==============================")
		fmt.Println(" ⚔️  WitchCraft - Batalha ")
		fmt.Println("==============================")
		fmt.Println("1️⃣  - Jogar Carta")
		fmt.Println("2️⃣  - Passar Turno")
		fmt.Println("3️⃣  - Atacar")
		fmt.Println("------------------------------")
		fmt.Print("👉 Escolha a sua ação de combate: ")

		var action int
		fmt.Scanln(&action)

		gameTurnMutex.RLock()
		turn := gameTurn
		gameTurnMutex.RUnlock()

		if turn != session_id {
			fmt.Println("❌ Ainda não é seu turno.")
		} else {
			switch action {
			case 1:
				play_card(encoder)
			case 0:
				fmt.Println("↩️ Voltando ao menu principal...")
				return
			default:
				fmt.Println("❌ Opção inválida. Tente novamente.")
			}
		}
	}
}

func prompt(prompt string) string {
	fmt.Printf("%s", prompt)
	var input string
	fmt.Scanln(&input)
	return input
}
