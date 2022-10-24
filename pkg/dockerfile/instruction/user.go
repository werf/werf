package instruction

type User struct {
	*Base

	User string
}

func NewUser(raw, user string) *User {
	return &User{Base: NewBase(raw), User: user}
}

func (i *User) Name() string {
	return "USER"
}
