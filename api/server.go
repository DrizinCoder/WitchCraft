package api

import (
	"WitchCraft/Cards"
	match "WitchCraft/Match"
	"WitchCraft/Player"
	"encoding/json"
	"fmt"
	"net"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

var playerManager = Player.NewManager()
var stock = Cards.NewStock()
var matchManager = match.NewMatchManager()

func Setup() {

	stock.CreateCard("Fireball", 10, 5, Cards.GOLD)
	stock.CreateCard("Icebolt", 8, 6, Cards.SILVER)
	stock.CreateCard("Goblin", 5, 10, Cards.BRONZE)
	stock.CreateCard("Dragon", 20, 20, Cards.DIAMOND)
	stock.CreateCard("Knight", 12, 15, Cards.SILVER)
	stock.CreateCard("Elf", 7, 8, Cards.BRONZE)

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
			fmt.Println("Erro ao conectar:", nil)
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
			continue
		}

		switch msg.Action {
		case "create_player":
			createPlayerHandler(msg, *encoder)
		}

	}
}

func createPlayerHandler(msg Message, encoder json.Encoder) {

	type req struct {
		Username string `json:"username"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var r req

	err := json.Unmarshal(msg.Data, &r)

	if err != nil {
		encoder.Encode(map[string]string{"error": err.Error()})
	}

	player := playerManager.Create_Player(r.Username, r.Login, r.Password)
	encoder.Encode(player)
}

func loginPlayerHlander() {

}

func openPackHandler() {

}

func getPlayerHandler() {

}

func enqueue() {

}

/*
---oq falta implementar---

-Tudo
*/
