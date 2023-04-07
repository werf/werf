package thirdparty

import "strings"

func MatchesAncestor(imgName, imgID, argName string) bool {
	if MatchesID(imgID, argName) {
		return true
	}
	return MatchesReference(imgName, argName)
}

func MatchesID(imageID, argID string) bool {
	return strings.HasPrefix(imageID, argID)
}

func MatchesReference(name, argName string) bool {
	if argName == "" {
		return true
	}
	splitName := strings.Split(name, ":")
	// If the arg contains a tag, we handle it differently than if it does not
	if strings.Contains(argName, ":") {
		splitArg := strings.Split(argName, ":")
		return strings.HasSuffix(splitName[0], splitArg[0]) && (splitName[1] == splitArg[1])
	}
	return strings.HasSuffix(splitName[0], argName)
}

func MatchesCtrName(ctrName, argName string) bool {
	return strings.Contains(ctrName, argName)
}
