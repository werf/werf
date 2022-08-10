package common

import (
	"fmt"
	"testing"
)

func Test_hashOriginUrl(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"git@github.com:werf/werf.git", "590dba0ea49212652f94e1cb1dbf3cc673d1dd964dea5d35239afcbbb3d38690"},

		{"git@github.com:user/repository.git", "67faac1e7c66486620847c62a962fe00bb1055a78567c81d4abd8b6ce1f201dc"},
		{"https://github.com/user/repository.git", "67faac1e7c66486620847c62a962fe00bb1055a78567c81d4abd8b6ce1f201dc"},
		{"https://github.com/user/repository", "67faac1e7c66486620847c62a962fe00bb1055a78567c81d4abd8b6ce1f201dc"},
		{"http://git:pass@github.com:9418/user/repository.git?foo#bar", "67faac1e7c66486620847c62a962fe00bb1055a78567c81d4abd8b6ce1f201dc"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test %v:", tt.arg), func(t *testing.T) {
			got, err := hashOriginUrl(tt.arg)
			if err != nil || got != tt.want {
				t.Errorf("hashOriginUrl() = %v, want %v, err %v", got, tt.want, err)
			}
		})
	}
}
