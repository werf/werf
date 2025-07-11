// Code generated by "enumer -type=StandardType -trimprefix=StandardType -linecomment -transform=lower -output=./standard_type_enumer.go"; DO NOT EDIT.

package sbom

import (
	"fmt"
	"strings"
)

const _StandardTypeName = "cyclonedx@1.6cyclonedx@1.5spdx@2.3spdx@2.2"

var _StandardTypeIndex = [...]uint8{0, 13, 26, 34, 42}

const _StandardTypeLowerName = "cyclonedx@1.6cyclonedx@1.5spdx@2.3spdx@2.2"

func (i StandardType) String() string {
	if i >= StandardType(len(_StandardTypeIndex)-1) {
		return fmt.Sprintf("StandardType(%d)", i)
	}
	return _StandardTypeName[_StandardTypeIndex[i]:_StandardTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _StandardTypeNoOp() {
	var x [1]struct{}
	_ = x[StandardTypeCycloneDX16-(0)]
	_ = x[StandardTypeCycloneDX15-(1)]
	_ = x[StandardTypeSPDX23-(2)]
	_ = x[StandardTypeSPDX22-(3)]
}

var _StandardTypeValues = []StandardType{StandardTypeCycloneDX16, StandardTypeCycloneDX15, StandardTypeSPDX23, StandardTypeSPDX22}

var _StandardTypeNameToValueMap = map[string]StandardType{
	_StandardTypeName[0:13]:       StandardTypeCycloneDX16,
	_StandardTypeLowerName[0:13]:  StandardTypeCycloneDX16,
	_StandardTypeName[13:26]:      StandardTypeCycloneDX15,
	_StandardTypeLowerName[13:26]: StandardTypeCycloneDX15,
	_StandardTypeName[26:34]:      StandardTypeSPDX23,
	_StandardTypeLowerName[26:34]: StandardTypeSPDX23,
	_StandardTypeName[34:42]:      StandardTypeSPDX22,
	_StandardTypeLowerName[34:42]: StandardTypeSPDX22,
}

var _StandardTypeNames = []string{
	_StandardTypeName[0:13],
	_StandardTypeName[13:26],
	_StandardTypeName[26:34],
	_StandardTypeName[34:42],
}

// StandardTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func StandardTypeString(s string) (StandardType, error) {
	if val, ok := _StandardTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _StandardTypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to StandardType values", s)
}

// StandardTypeValues returns all values of the enum
func StandardTypeValues() []StandardType {
	return _StandardTypeValues
}

// StandardTypeStrings returns a slice of all String values of the enum
func StandardTypeStrings() []string {
	strs := make([]string, len(_StandardTypeNames))
	copy(strs, _StandardTypeNames)
	return strs
}

// IsAStandardType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i StandardType) IsAStandardType() bool {
	for _, v := range _StandardTypeValues {
		if i == v {
			return true
		}
	}
	return false
}
