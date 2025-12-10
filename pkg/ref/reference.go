/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ref // import "helm.sh/helm/v3/internal/experimental/registry"

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
)

type (
	// Reference defines the main components of a reference specification
	Reference struct {
		Tag  string
		Repo string
	}
)

// ParseReference converts a string to a Reference
func ParseReference(s string) (*Reference, error) {
	r, err := name.ParseReference(s, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	return &Reference{
		Tag:  r.Identifier(),
		Repo: r.Context().Name(),
	}, nil
}

// FullName the full name of a reference (repo:tag)
func (ref *Reference) FullName() string {
	if ref.Tag == "" {
		return ref.Repo
	}
	return fmt.Sprintf("%s:%s", ref.Repo, ref.Tag)
}
