package api

import (
	"encoding/json"
	"fmt"
	"net"
)

func StartUDPServer(addr string) {
	conn, err := net.ListenPacket("udp", addr)
	fmt.Println("Servidor UDP rodando em", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		var msg Message
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}

		if msg.Action == "ping" {
			resp := Message{Action: "pong", Data: []byte("{}")}
			resp_json, _ := json.Marshal(resp)
			conn.WriteTo(resp_json, clientAddr)
		}
	}
}
