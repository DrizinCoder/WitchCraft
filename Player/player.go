package Player

import "WitchCraft/Cards"

type Player struct {
	ID       int
	UserName string
	Login    string
	Password string
	Cards    []*Cards.Card
}

func New_Player(id int, userName string, login string, password string) *Player {
	return &Player{
		ID:       id,
		UserName: userName,
		Login:    login,
		Password: password,
		Cards:    make([]*Cards.Card, 0),
	}
}
