package storage

type DeleteImageOptions struct {
	RmiForce                 bool
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
}
