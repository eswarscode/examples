package pkg2

type Player struct {
	Username string
	Level    int
}

func NewPlayer(username string, level int) *Player {
	return &Player{Username: username, Level: level}
}
