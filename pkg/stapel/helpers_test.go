package stapel

import (
	dockercontainer "github.com/docker/docker/api/types/container"
)

func stapelVolumeMount(volumeName string) dockercontainer.MountPoint {
	return dockercontainer.MountPoint{
		Type:        "volume",
		Name:        volumeName,
		Destination: containerVolumeDestination,
	}
}

// fakeContainer is an inspect response as the daemon would report it for a
// container created from imageRef under name. The reported image is the id the
// daemon resolves the reference to, which differs from the creation reference
// of Config as soon as the tag is moved to another image.
func fakeContainer(id, name, imageRef string, mounts ...dockercontainer.MountPoint) dockercontainer.InspectResponse {
	return dockercontainer.InspectResponse{
		ContainerJSONBase: &dockercontainer.ContainerJSONBase{
			ID:    id,
			Name:  "/" + name,
			Image: "sha256:" + id,
		},
		Mounts: mounts,
		Config: &dockercontainer.Config{Image: imageRef},
	}
}
