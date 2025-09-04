package main

// Apenas testes até o momento
import (
	client "WitchCraft/Client"
	"WitchCraft/api"
	"WitchCraft/stress"
	"fmt"
	"os"
)

func main() {
	mode := os.Getenv("MODE")

	switch mode {
	case "server":
		fmt.Println("Iniciando servidor WitchCraft...")
		go api.StartUDPServer(":9999")
		api.Setup()
	case "client":
		fmt.Println("Iniciando cliente WitchCraft...")
		client.Setup()
	case "stress":
		fmt.Println("Iniciando teste de estresse...")
		stress.Run()
	default:
		fmt.Println("Defina a variável MODE com 'server' ou 'client'")
	}
}
