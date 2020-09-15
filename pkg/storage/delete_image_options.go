package storage

type DeleteImageOptions struct {
	RmiForce bool
}

type FilterStagesAndProcessRelatedDataOptions struct {
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
}
