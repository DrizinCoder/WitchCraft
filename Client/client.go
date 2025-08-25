package client

import (
	"WitchCraft/Cards"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
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

var session_id int
var start time.Time
var channel chan int

func Setup() {

	serverAddr := os.Getenv("SERVER_ADDR")
	conn, err := net.Dial("tcp", serverAddr)

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	go handleConnection(decoder)

	channel = make(chan int, 1)

	for {

		select {
		case change := <-channel:
			if change == 1 {
				GameMenu()
			}
		default:

		}

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
		fmt.Scanln(&action)

		switch action {
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
			seeInventory(encoder)
		case 7:
			ping(encoder)
		case 0:
			fmt.Println("👋 Saindo do jogo... Até logo!")
			return
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
		case "Game_start":
			channel <- 1
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
		return
	}
	payload := Req_id{
		ID: session_id,
	}

	sendRequest(encoder, "enqueue_player", payload)
}

func seeInventory(encoder *json.Encoder) {
	if session_id == 0 {
		return
	}
	payload := Req_id{
		ID: session_id,
	}

	sendRequest(encoder, "see_inventory", payload)
}

func ping(encoder *json.Encoder) {
	start = time.Now()

	payload := ""
	sendRequest(encoder, "ping", payload)
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
}

func handleOpenPackResponse(data json.RawMessage) {

	var cards []Cards.Card
	err := json.Unmarshal(data, &cards)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados: ", err)
		return
	}

	fmt.Println("Pacote Aberto. Cartas: ")
	for _, c := range cards {
		fmt.Printf("- %s (Power: %d, Life: %d, Rarity: %s)\n",
			c.Name, c.Power, c.Life, c.Rarity)
	}
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

	fmt.Println("Inventário de cartas: ")
	if len(cards) == 0 {
		fmt.Println("Sem cartas no inventário.")
	} else {
		for _, c := range cards {
			fmt.Printf("- %s (Power: %d, Life: %d, Rarity: %s)\n",
				c.Name, c.Power, c.Life, c.Rarity)
		}
	}
}

func handlePongResponse() {
	elapsed := time.Since(start)
	fmt.Printf("Ping: %s\n", elapsed)
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

func GameMenu() {
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

		switch action {
		case 1:
			fmt.Println("🃏 Você escolheu Jogar uma Carta.")
			// Aqui você pode chamar uma função como `playCard()`
		case 2:
			fmt.Println("⏭️ Você passou o turno.")
			// Aqui você pode chamar algo como `passTurn()`
		case 3:
			fmt.Println("⚔️ Você escolheu Atacar.")
			// Aqui você pode chamar uma função como `attack()`
		case 0:
			fmt.Println("↩️ Voltando ao menu principal...")
			return
		default:
			fmt.Println("❌ Opção inválida. Tente novamente.")
		}
	}
}

func prompt(prompt string) string {
	fmt.Printf("%s", prompt)
	var input string
	fmt.Scanln(&input)
	return input
}
