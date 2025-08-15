package main

import "fmt"

const (
	IRON = iota
	BRONZE
	SILVER
	GOLD
	DIAMOND
)

type Rare uint8

type Card struct {
	ID     int
	Name   string
	Power  int
	Life   int
	Rarity Rare
}

func New_Card(id int, name string, power int, life int, rarity Rare) *Card {
	return &Card{
		ID:     id,
		Name:   name,
		Power:  power,
		Life:   life,
		Rarity: rarity,
	}
}

func (r Rare) String() string {
	switch r {
	case IRON:
		return "iron"
	case SILVER:
		return "silver"
	case GOLD:
		return "gold"
	case DIAMOND:
		return "diamond"
	case BRONZE:
		return "bronze"
	default:
		return "unknown"
	}
}

func (c *Card) Hit() {
	fmt.Printf("A carta %s deu %d de dano\nSua Raridade é: %s\n", c.Name, c.Power, c.Rarity)
}

func main() {
	p := New_Card(1, "The Hunter", 50, 100, GOLD)
	p.Hit()
	a := New_Card(1, "Smoker", 70, 50, DIAMOND)
	a.Hit()
}
