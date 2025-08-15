package Cards

import (
	"errors"
	"sync"
)

type Stock struct {
	mu   sync.Mutex
	Deck []*Card
}

var nextID int
var muID sync.Mutex

func generateID() int {
	muID.Lock()
	defer muID.Unlock()
	nextID++
	return nextID
}

func NewStock() *Stock {
	return &Stock{
		Deck: make([]*Card, 0),
	}
}

func (s *Stock) CreateCard(name string, power int, life int, rarity Rare) *Card {
	s.mu.Lock()
	defer s.mu.Unlock()

	card := New_Card(generateID(), name, power, life, rarity)
	s.Deck = append(s.Deck, card)

	return card
}

func (s *Stock) RemoveCard(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, c := range s.Deck {
		if c.ID == id {
			s.Deck[i] = s.Deck[len(s.Deck)-1]
			s.Deck = s.Deck[:len(s.Deck)-1]
			return nil
		}
	}

	return errors.New("Card not found")
}

func (s *Stock) GeneratePack() ([]*Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Deck) < 5 {
		return nil, errors.New("não há cartas suficientes no estoque")
	}

	pack := make([]*Card, 0, 5)
	toRemove := s.Deck[:5]

	for _, c := range toRemove {
		pack = append(pack, c)
	}

	s.Deck = s.Deck[5:]

	return pack, nil
}
