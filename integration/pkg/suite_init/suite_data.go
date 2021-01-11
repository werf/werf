package suite_init

type SuiteData struct {
	*StubsData
	*SynchronizedSuiteCallbacksData
	*WerfBinaryData
	*ProjectNameData
	*K8sDockerRegistryData
	*TmpDirData
	*ContainerRegistryPerImplementationData
}

func (data *SuiteData) SetupStubs(setupData *StubsData) bool {
	data.StubsData = setupData
	return true
}

func (data *SuiteData) SetupSynchronizedSuiteCallbacks(setupData *SynchronizedSuiteCallbacksData) bool {
	data.SynchronizedSuiteCallbacksData = setupData
	return true
}

func (data *SuiteData) SetupWerfBinary(setupData *WerfBinaryData) bool {
	data.WerfBinaryData = setupData
	return true
}

func (data *SuiteData) SetupProjectName(setupData *ProjectNameData) bool {
	data.ProjectNameData = setupData
	return true
}

func (data *SuiteData) SetupK8sDockerRegistry(setupData *K8sDockerRegistryData) bool {
	data.K8sDockerRegistryData = setupData
	return true
}

func (data *SuiteData) SetupTmp(setupData *TmpDirData) bool {
	data.TmpDirData = setupData
	return true
}

func (data *SuiteData) SetupContainerRegistryPerImplementation(setupData *ContainerRegistryPerImplementationData) bool {
	data.ContainerRegistryPerImplementationData = setupData
	return true
}
