package Player

import (
	"WitchCraft/Cards"
	"errors"
	"net"
	"sync"
)

type Manager struct {
	mu             sync.Mutex
	PlayersByID    map[int]*Player
	PlayersByLogin map[string]*Player
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
		PlayersByID:    make(map[int]*Player),
		PlayersByLogin: make(map[string]*Player),
	}
}

func (m *Manager) Create_Player(name string, login string, password string) (*Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" || login == "" || password == "" {
		return nil, errors.New("name, login or password cannot be blank space")
	}

	_, exists := m.PlayersByLogin[login]
	if exists {
		return nil, errors.New("login already exists")
	}

	player := New_Player(generateID(), name, login, password)
	m.PlayersByID[player.ID] = player
	m.PlayersByLogin[login] = player

	return player, nil
}

func (m *Manager) Login(login string, password string, conn net.Conn) (*Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByLogin[login]
	if !exists || player.Password != password {
		return nil, errors.New("invalid credentials")
	}

	player.Conn = conn
	return player, nil
}

func (m *Manager) Search_Player_ByID(id int) (*Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return player, nil
}

func (m *Manager) Open_pack(PlayerId int, stock *Cards.Stock) ([]*Cards.Card, error) {

	pack, err := stock.GeneratePack()
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByID[PlayerId]
	if !exists {
		return nil, errors.New("user not found")
	}

	player.Cards = append(player.Cards, pack...)
	return pack, nil
}

func (m *Manager) Search_Player_ByLogin(login string) (*Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByLogin[login]
	if !exists {
		return nil, errors.New("user not found")
	}
	return player, nil
}

func (m *Manager) Get_inventory(PlayerID int) ([]*Cards.Card, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByID[PlayerID]
	if !exists {
		return nil, errors.New("user not found")
	}

	result := make([]*Cards.Card, len(player.Cards))
	copy(result, player.Cards)
	return result, nil
}

func (m *Manager) Get_deck(PlayerID int) ([]*Cards.Card, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByID[PlayerID]
	if !exists {
		return nil, errors.New("user not found")
	}

	result := make([]*Cards.Card, len(player.GameDeck))
	copy(result, player.GameDeck)
	return result, nil
}

func (m *Manager) SetDeck(playerID int, deck []*Cards.Card) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.PlayersByID[playerID]
	if !exists {
		return errors.New("user not found")
	}

	type cardKey struct {
		Name        string
		Power       int
		Life        int
		Inteligence int
		Rarity      Cards.Rare
	}

	invSet := make(map[cardKey]struct{}, len(player.Cards))
	for _, c := range player.Cards {
		invSet[cardKey{c.Name, c.Power, c.Life, c.Inteligence, c.Rarity}] = struct{}{}
	}

	for _, dc := range deck {
		key := cardKey{dc.Name, dc.Power, dc.Life, dc.Inteligence, dc.Rarity}
		if _, ok := invSet[key]; !ok {
			return errors.New("uma ou mais cartas do deck não estão no inventário")
		}
	}

	player.GameDeck = deck
	return nil
}
