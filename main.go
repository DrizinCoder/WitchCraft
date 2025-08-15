package main

// Apenas testes até o momento

import (
	"WitchCraft/Cards"
	"fmt"
)

func main() {
	s := Cards.NewStock()
	// Cartas principais
	s.CreateCard("The Hunter", 50, 100, Cards.GOLD)
	s.CreateCard("Smoker", 70, 80, Cards.SILVER)
	s.CreateCard("Chainner", 40, 120, Cards.BRONZE)
	s.CreateCard("Occult", 60, 90, Cards.DIAMOND)
	s.CreateCard("Yagorath", 100, 150, Cards.DIAMOND)

	// Cartas adicionais
	s.CreateCard("Fire Wraith", 80, 60, Cards.GOLD)
	s.CreateCard("Shadow Monk", 35, 110, Cards.BRONZE)
	s.CreateCard("Venom Fang", 65, 85, Cards.SILVER)
	s.CreateCard("Titan Guard", 45, 140, Cards.IRON)
	s.CreateCard("Blood Reaper", 90, 70, Cards.GOLD)
	s.CreateCard("Ice Whisper", 55, 100, Cards.SILVER)
	s.CreateCard("Arcane Sentry", 40, 130, Cards.BRONZE)
	s.CreateCard("Night Howler", 75, 95, Cards.GOLD)
	s.CreateCard("Blight Priest", 60, 120, Cards.DIAMOND)
	s.CreateCard("Ironborn", 30, 150, Cards.IRON)
	s.CreateCard("Skull Crusher", 95, 60, Cards.GOLD)
	s.CreateCard("Storm Caller", 70, 90, Cards.SILVER)
	s.CreateCard("Flesh Golem", 50, 160, Cards.BRONZE)
	s.CreateCard("Soul Harvester", 85, 100, Cards.DIAMOND)
	s.CreateCard("Ash Prophet", 65, 110, Cards.SILVER)

	exibir(s.Deck)

	fmt.Println("-----------------------------------------------------------------------")

	pack, _ := s.GeneratePack()

	fmt.Println("=-----------------------------------------------------------------------=")

	exibir(pack)

	fmt.Println("-----------------------------------------------------------------------")

	exibir(s.Deck)
}

func exibir(s []*Cards.Card) {
	for _, card := range s {
		fmt.Printf("Carta criada: %s (Power: %d, Life: %d, Rarity: %s)\n",
			card.Name, card.Power, card.Life, card.Rarity)
	}
}
