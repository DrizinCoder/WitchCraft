package main

// Apenas testes até o momento
import (
	client "WitchCraft/Client"
	"WitchCraft/api"
	"fmt"
)

func main() {
	var a int
	fmt.Scanln(&a)

	switch a {
	case 1:
		api.Setup()
	case 2:
		client.Setup()
	}
}
