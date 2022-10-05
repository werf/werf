package dockerfile_instruction

type DockerfileInstruction interface {
	Name() string
}
