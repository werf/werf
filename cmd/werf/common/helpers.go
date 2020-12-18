package common

import "time"

func NewDuration(value time.Duration) *time.Duration {
	if value != 0 {
		res := new(time.Duration)
		*res = value
		return res
	}
	return nil
}

func NewBool(value bool) *bool {
	res := new(bool)
	*res = value
	return res
}
