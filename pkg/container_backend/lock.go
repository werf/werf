package container_backend

import "fmt"

func ContainerLockName(containerName string) string {
	return fmt.Sprintf("container.%s", containerName)
}

func ImageLockName(imageName string) string {
	return fmt.Sprintf("image.%s", imageName)
}
