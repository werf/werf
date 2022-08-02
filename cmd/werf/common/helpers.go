package common

import "time"

func NewDuration(value time.Duration) *time.Duration {
	res := new(time.Duration)
	*res = value
	return res
}

func NewBool(value bool) *bool {
	res := new(bool)
	*res = value
	return res
}
