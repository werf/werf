package container_runtime

type ImageInterface interface {
	SetBuiltID(builtID string)
	GetBuiltID() string
}
