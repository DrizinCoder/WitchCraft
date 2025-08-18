package client

import (
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

	go handleConnection(decoder)

	select {}

}

func handleConnection(decoder json.Decoder) {

}

func handleCreatePlayerResponse() {

}

func handleLoginPlayerResponse() {

}

func handleOpenPackResponse() {

}

func handleSearchPlayerResponse() {

}

func handleEnqueueResponse() {

}
