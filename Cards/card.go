package cards

import "fmt"

type Card struct {
	ID    int
	Name  string
	Power int
	Life  int
}

func New_Card(name string, power int, life int) *Card {
	return &Card{
		Name:  name,
		Power: power,
		Life:  life,
	}
}

func (c *Card) Hit() {
	fmt.Printf("A carta %s deu %d de dano\n", c.Name, c.Power)
}

func main() {
	p := New_Card("The Hunter", 50, 100)
	p.Hit()
	a := New_Card("Smoker", 70, 50)
	a.Hit()
}
