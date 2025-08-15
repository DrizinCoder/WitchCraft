package player

type Player struct {
	ID       int
	UserName string
	Login    string
	Password string
}

func New_Player(id int, userName string, login string, password string) *Player {
	return &Player{
		ID:       id,
		UserName: userName,
		Login:    login,
		Password: password,
	}
}
