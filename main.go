package main

// Apenas testes até o momento
import (
	client "WitchCraft/Client"
	"WitchCraft/api"
	"fmt"
	"os"
)

func main() {
	mode := os.Getenv("MODE") // lê variável de ambiente

	switch mode {
	case "server":
		fmt.Println("Iniciando servidor WitchCraft...")
		api.Setup()
	case "client":
		fmt.Println("Iniciando cliente WitchCraft...")
		client.Setup()
	default:
		fmt.Println("Defina a variável MODE com 'server' ou 'client'")
	}
}
