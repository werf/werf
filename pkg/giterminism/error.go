package giterminism

type ConfigNotFoundError error

func IsConfigNotFoundError(err error) bool {
	switch err.(type) {
	case ConfigNotFoundError:
		return true
	default:
		return false
	}
}
