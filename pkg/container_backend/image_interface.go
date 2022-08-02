package container_backend

type ImageInterface interface {
	Name() string
	SetName(name string)

	SetBuiltID(builtID string)
	BuiltID() string
}
