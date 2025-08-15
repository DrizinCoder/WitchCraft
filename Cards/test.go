package Cards

import (
	"fmt"
)

func main() {
	s := NewStock()

	// Cartas principais
	s.CreateCard("The Hunter", 50, 100, GOLD)
	s.CreateCard("Smoker", 70, 80, SILVER)
	s.CreateCard("Chainner", 40, 120, BRONZE)
	s.CreateCard("Occult", 60, 90, DIAMOND)
	s.CreateCard("Yagorath", 100, 150, DIAMOND)

	// Cartas adicionais
	s.CreateCard("Fire Wraith", 80, 60, GOLD)
	s.CreateCard("Shadow Monk", 35, 110, BRONZE)
	s.CreateCard("Venom Fang", 65, 85, SILVER)
	s.CreateCard("Titan Guard", 45, 140, IRON)
	s.CreateCard("Blood Reaper", 90, 70, GOLD)
	s.CreateCard("Ice Whisper", 55, 100, SILVER)
	s.CreateCard("Arcane Sentry", 40, 130, BRONZE)
	s.CreateCard("Night Howler", 75, 95, GOLD)
	s.CreateCard("Blight Priest", 60, 120, DIAMOND)
	s.CreateCard("Ironborn", 30, 150, IRON)
	s.CreateCard("Skull Crusher", 95, 60, GOLD)
	s.CreateCard("Storm Caller", 70, 90, SILVER)
	s.CreateCard("Flesh Golem", 50, 160, BRONZE)
	s.CreateCard("Soul Harvester", 85, 100, DIAMOND)
	s.CreateCard("Ash Prophet", 65, 110, SILVER)

	// Exibir todas as cartas criadas
	for _, card := range s.Deck {
		fmt.Printf("Carta criada: %s (Power: %d, Life: %d, Rarity: %s)\n",
			card.Name, card.Power, card.Life, card.Rarity)
	}
}
