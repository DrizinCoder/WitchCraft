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

func setup() {

	conn, err := net.Dial("tcp", ":8080")

	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	go handleConnection(*decoder)

	select {}

}

func handleConnection(decoder json.Decoder) {

	for {
		var payload Message
		err := decoder.Decode(payload)
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
	type req struct {
		Pack []*Cards.Card `json:"pack"`
	}

	var resp req
	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
		return
	}

	fmt.Println("Pacote Aberto. Cartas: ")
	fmt.Println("ID:", resp.Pack)
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
		msg map[string]string `json:"payload"`
	}

	var resp req

	err := json.Unmarshal(data, &resp)

	if err != nil {
		fmt.Println("Erro ao decodificar pacote de dados")
	}

	fmt.Println(resp.msg)
}
