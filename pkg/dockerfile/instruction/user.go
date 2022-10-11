package instruction

type User struct {
	User string
}

func NewUser(user string) *User {
	return &User{User: user}
}

func (i *User) Name() string {
	return "USER"
}
