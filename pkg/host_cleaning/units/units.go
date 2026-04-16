package units

/*
Package units implements the UnitValue type, which provides a unified way to handle
storage thresholds in CLI flags. It supports both percentages and absolute sizes (e.g., GB, MiB).

Final Specification (SDD):

| Category        | Scenario                             | Input Examples                                 | Result / Logic                                                 |
|-----------------|--------------------------------------|------------------------------------------------|----------------------------------------------------------------|
| Simple Units    | Using percentages (Legacy)           | usage: "70", margin: "10"                      | Parsed as percentages. Threshold = (Total * 70) / 100.         |
| Simple Units    | Using absolute sizes                 | usage: "50GB", margin: "5GB"                   | Parsed as bytes. Threshold = ValueInBytes.                     |
| Combined Units  | Different units for DIFFERENT groups | backend-usage: "70", local-cache-usage: "10GB" | VALID. Each group is independent.                              |
| Mixed Units     | Different units within SAME group    | backend-usage: "100GB", backend-margin: "5"    | VALIDATION ERROR. Unit consistency is required within a group. |

Parsing behavior and validation notes:
- Parsing Priority: Input is ALWAYS attempted to be parsed as a percentage (pure number) first
  to maintain backward compatibility.
- Error Reporting: If both percentage and byte-size parsing fail, a detailed error message
  including underlying errors from both parsers is returned to the user.
- Constraints: Percentage values must be within the [0, 100] range. Valid byte units
  include decimal (MB, GB, TB) and binary (MiB, GiB, TiB) prefixes.
*/

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/go-units"
)

type UnitValue struct {
	Value   uint64
	IsBytes bool
}

func (s *UnitValue) Set(input string) error {
	v, err := ParseUnitValue(input)
	if err != nil {
		return err
	}
	*s = *v
	return nil
}

// Type is required to satisfy the pflag.Value interface used by cobra flags.
// It defines the value type displayed in the CLI help message.
func (s *UnitValue) Type() string {
	return "unitValue"
}

// ParseUnitValue parses the input string into a UnitValue.
// It first attempts to parse the input as a percentage (a pure number between 0 and 100).
// If that fails, it tries to parse it as an absolute storage size (e.g., "10GB", "500MiB").
// The order is important to maintain backward compatibility for pure numeric inputs.
func ParseUnitValue(input string) (*UnitValue, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty storage value")
	}

	val, pctErr := strconv.ParseUint(input, 10, 64)
	if pctErr == nil {
		if val > 100 {
			return nil, fmt.Errorf("percentage value %d cannot exceed 100", val)
		}
		return &UnitValue{Value: val, IsBytes: false}, nil
	}

	bytes, bytesErr := units.RAMInBytes(input)
	if bytesErr == nil {
		return &UnitValue{Value: uint64(bytes), IsBytes: true}, nil
	}

	return nil, fmt.Errorf("invalid storage value %q: specify percentage (0-100, error: %v) or absolute size (e.g. 10GB, 500MiB, error: %v)", input, pctErr, bytesErr)
}

func (s *UnitValue) ToBytes(totalBytes uint64) uint64 {
	if s.IsBytes {
		return s.Value
	}

	return (totalBytes * s.Value) / 100
}

func (s *UnitValue) String() string {
	if s.IsBytes {
		return units.BytesSize(float64(s.Value))
	}
	return fmt.Sprintf("%d", s.Value)
}
