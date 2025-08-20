package client

import (
	"WitchCraft/Cards"
	"encoding/json"
	"fmt"
	"net"
	"os"
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

	for {
		fmt.Println("\n==============================")
		fmt.Println(" 🎮 WitchCraft - Menu Principal ")
		fmt.Println("==============================")
		fmt.Println("1️⃣  - Registrar Jogador")
		fmt.Println("2️⃣  - Login")
		fmt.Println("3️⃣  - Abrir Pacote de Cartas")
		fmt.Println("4️⃣  - Buscar Jogador")
		fmt.Println("5️⃣  - Entrar na Fila")
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
		}
	}

}

func createPlayer(encoder *json.Encoder) {

	var username, login, password string

	fmt.Print("👤 Digite o nome do jogador: ")
	fmt.Scanln(&username)

	fmt.Print("📧 Digite o login: ")
	fmt.Scanln(&login)

	fmt.Print("🔑 Digite a senha: ")
	fmt.Scanln(&password)

	payload := Req_player{
		Username: username,
		Login:    login,
		Password: password,
	}

	data, _ := json.Marshal(payload)

	req := Message{
		Action: "create_player",
		Data:   data,
	}

	encoder.Encode(req)

}

func loginPlayer(encoder *json.Encoder) {

	var login, password string

	fmt.Print("📧 Digite o login: ")
	fmt.Scanln(&login)

	fmt.Print("🔑 Digite a senha: ")
	fmt.Scanln(&password)

	payload := Req_login{
		Login:    login,
		Password: password,
	}

	data, _ := json.Marshal(payload)

	req := Message{
		Action: "login_player",
		Data:   data,
	}

	encoder.Encode(req)

}

func openPack(encoder *json.Encoder) {
	if session_id == 0 {
		return
	}

	payload := Req_id{
		ID: session_id,
	}

	println(payload.ID)

	data, _ := json.Marshal(payload)

	req := Message{
		Action: "open_pack",
		Data:   data,
	}

	encoder.Encode(req)

}

func searchPlayer(encoder *json.Encoder) {
	if session_id == 0 {
		return
	}

	payload := Req_id{
		ID: session_id,
	}

	println(payload.ID)

	data, _ := json.Marshal(payload)

	req := Message{
		Action: "search_player",
		Data:   data,
	}

	encoder.Encode(req)
}

func enqueue(encoder *json.Encoder) {
	if session_id == 0 {
		return
	}
	payload := Req_id{
		ID: session_id,
	}

	println(payload.ID)

	data, _ := json.Marshal(payload)

	req := Message{
		Action: "enqueue_player",
		Data:   data,
	}

	encoder.Encode(req)
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
