package main

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

	pack := make([]*Card, 5)
	for _, c := range s.Deck[:5] {
		pack = append(pack, c)
		s.RemoveCard(c.ID)
	}

	return pack, nil
}
