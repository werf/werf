package util

import (
	"fmt"
	"testing"
)

func TestReverse(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"", ""},
		{"Hello, world", "dlrow ,olleH"},
		{"Hello, 世界", "界世 ,olleH"},
		{"😍👀", "👀😍"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %v:", i), func(t *testing.T) {
			if got := Reverse(tt.arg); got != tt.want {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}
