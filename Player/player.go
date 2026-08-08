package Player

import (
	"WitchCraft/Cards"
	"encoding/json"
	"errors"
	"net"
	"sync"
)

type Player struct {
	ID       int
	UserName string
	Login    string
	Password string
	Cards    []*Cards.Card
	GameDeck []*Cards.Card
	In_game  bool
	Conn     net.Conn
	mu       sync.Mutex
}

func New_Player(id int, userName string, login string, password string) *Player {
	return &Player{
		ID:       id,
		UserName: userName,
		Login:    login,
		Password: password,
		Cards:    make([]*Cards.Card, 0),
		GameDeck: make([]*Cards.Card, 0),
		In_game:  false,
	}
}

func (p *Player) Send(v any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Conn == nil {
		return errors.New("player has no active connection")
	}
	return json.NewEncoder(p.Conn).Encode(v)
}
