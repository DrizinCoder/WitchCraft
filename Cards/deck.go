package main

import "sync"

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
