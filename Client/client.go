package client

import (
	"WitchCraft/Cards"
	"encoding/json"
	"fmt"
	"net"
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

func Setup() {

	conn, err := net.Dial("tcp", ":8080")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	go handleConnection(decoder)

	for {

		var action int
		fmt.Println("Escolha a sua próxima ação.")
		fmt.Printf("1 - Register Player\n2 - Login\n3 - Search Player\n4 - Open Pack\n5 - Enqueue\n")
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
		default:
			fmt.Println("Unknown value.")
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
		}
	}

}

func createPlayer(encoder *json.Encoder) {

	var username string
	var login string
	var password string

	fmt.Scanf("%s %s %s", &username, &login, &password)

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

	var login string
	var password string

	fmt.Scanf("%s %s", &login, &password)

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

	var id int

	fmt.Scanf("%d", &id)

	payload := Req_id{
		ID: id,
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

}

func enqueue(encoder *json.Encoder) {

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
	type req struct {
		Msg map[string]string `json:"payload"`
	}

	var resp req

	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
	}

	fmt.Println(resp.Msg)
}
