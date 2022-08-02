package utils

import "strings"

func GetBuiltImageLastStageImageName(testDirPath, werfBinPath, imageName string) string {
	stageImageName := SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		"stage", "image", imageName,
	)

	return strings.TrimSpace(stageImageName)
}
