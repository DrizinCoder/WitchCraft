package player

import (
	"errors"
	"sync"
)

type Manager struct {
	mu      sync.Mutex
	Players []*Player
}

var nextID int
var muID sync.Mutex

func generateID() int {
	muID.Lock()
	defer muID.Unlock()
	nextID++
	return nextID
}

func NewManager() *Manager {
	return &Manager{
		Players: make([]*Player, 0),
	}
}

func (m *Manager) Create_Player(name string, login string, password string) *Player {
	m.mu.Lock()
	defer m.mu.Unlock()

	player := New_Player(generateID(), name, login, password)
	m.Players = append(m.Players, player)

	return player
}

func (m *Manager) Search_Player(id int) (*Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.Players {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New("User not found")
}
