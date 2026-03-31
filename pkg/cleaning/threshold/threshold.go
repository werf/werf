package threshold

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

type Type string

const (
	TypePercentage Type = "percentage"
	TypeBytes      Type = "bytes"
)

type Threshold struct {
	Type  Type
	Value uint64
}

func NewPercentage(value uint64) Threshold {
	return Threshold{Type: TypePercentage, Value: value}
}

func NewBytes(value uint64) Threshold {
	return Threshold{Type: TypeBytes, Value: value}
}

func Parse(value string) (Threshold, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Threshold{}, fmt.Errorf("empty volume usage threshold")
	}

	if number, err := strconv.ParseUint(value, 10, 64); err == nil {
		if number > 100 {
			return Threshold{}, fmt.Errorf("percentage threshold %q should be between 0 and 100", value)
		}
		return NewPercentage(number), nil
	}

	bytesValue, err := humanize.ParseBytes(value)
	if err != nil {
		return Threshold{}, fmt.Errorf("parse volume usage threshold %q: %w", value, err)
	}
	if bytesValue == 0 {
		return Threshold{}, fmt.Errorf("bytes threshold %q should be greater than 0", value)
	}

	return NewBytes(bytesValue), nil
}

func (t Threshold) String() string {
	switch t.Type {
	case TypePercentage:
		return fmt.Sprintf("%d", t.Value)
	case TypeBytes:
		return humanize.Bytes(t.Value)
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", t.Type))
	}
}

func (t Threshold) FormatCLIValue() string {
	switch t.Type {
	case TypePercentage:
		return fmt.Sprintf("%d", t.Value)
	case TypeBytes:
		return fmt.Sprintf("%dB", t.Value)
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", t.Type))
	}
}

func (t Threshold) PercentageValue() float64 {
	if t.Type != TypePercentage {
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", t.Type))
	}
	return float64(t.Value)
}

func ValueOrDefault(optionValue *Threshold, defaultValue Threshold) Threshold {
	if optionValue != nil {
		return *optionValue
	}
	return defaultValue
}

func implicitBytesMargin(threshold, defaultMargin Threshold) Threshold {
	if threshold.Type != TypeBytes {
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", threshold.Type))
	}
	if defaultMargin.Type != TypePercentage {
		panic(fmt.Sprintf("unexpected default margin type %q", defaultMargin.Type))
	}

	return NewBytes((threshold.Value / 100) * defaultMargin.Value)
}

func Resolve(thresholdOption, marginOption *Threshold, defaultThreshold, defaultMargin Threshold, marginExplicit bool, thresholdFlagName, marginFlagName string) (Threshold, Threshold, error) {
	threshold := ValueOrDefault(thresholdOption, defaultThreshold)

	if !marginExplicit || marginOption == nil {
		if threshold.Type == TypeBytes {
			return threshold, implicitBytesMargin(threshold, defaultMargin), nil
		}
		return threshold, defaultMargin, nil
	}

	margin := *marginOption
	if threshold.Type != margin.Type {
		return Threshold{}, Threshold{}, fmt.Errorf("%s and %s must use the same format when both are set explicitly: both percentages (e.g. 70 and 5) or both with units (e.g. 10GB and 2GB)", thresholdFlagName, marginFlagName)
	}

	return threshold, margin, nil
}
