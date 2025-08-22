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

	player, err := m.Search_Player_ByID(PlayerId)
	if err != nil {
		return nil, err
	}

	pack, err := stock.GeneratePack()
	if err != nil {
		return nil, err
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

	player, exists := m.Search_Player_ByID(PlayerID)
	if exists != nil {
		return nil, errors.New("user not found")
	}

	return player.Cards, nil
}
