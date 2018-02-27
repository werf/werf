package config

import (
	"testing"
)

func Test_SplitDimgs_UnmarshalYAML_MustSetParents(t *testing.T) {
	data := `
dimg: app
git:
- add: '/app'
  to: '/'
  stageDependencies:
    install:
    - /src/main.go
shell:
  beforeInstall:
  - apt get update
mount:
- from: tmp_dir
  to: '/var/tmp'
docker:
  LABEL: {asd: qwe}
import:
- add: '/build/output'
  to: '/app/bin'
`

	docs := []*Doc{
		{
			Content:        []byte(data),
			Line:           1,
			RenderFilePath: "data",
		},
	}

	dimgs, err := splitByRawDimgs(docs)

	if err != nil {
		t.Fatalf("error occured: %v\n", err)
	}

	if len(dimgs) != 1 {
		t.Fatalf("dimgs len should be 1")
	}

	dimg := dimgs[0]

	if dimg.Doc == nil {
		t.Fatalf("dimg.Doc should be set\n")
	}

	// git:
	if len(dimg.RawGit) != 1 {
		t.Fatalf("dimg.RawGit len must be 1")
	}

	if dimg.RawGit[0].RawDimg == nil {
		t.Fatalf("dimg.RawGit[0].RawDimg should be set\n")
	}

	if dimg.RawGit[0].RawStageDependencies == nil {
		t.Fatalf("dimg.RawGit[0].RawStageDependencies must be set")
	}

	if dimg.RawGit[0].RawStageDependencies.RawGit == nil {
		t.Fatalf("dimg.RawGit[0].RawStageDependencies.RawShell should be set\n")
	}

	// shell:
	if dimg.RawShell.RawDimg == nil {
		t.Fatalf("dimg.RawShell.RawDimg should be set\n")
	}

	// mount:
	if len(dimg.RawMount) != 1 {
		t.Fatalf("dimg.RawMount len must be 1")
	}

	if dimg.RawMount[0].RawDimg == nil {
		t.Fatalf("dimg.RawMount[0].RawDimg should be set\n")
	}

	// docker:
	if dimg.RawDocker.RawDimg == nil {
		t.Fatalf("dimg.RawDocker.RawDimg should be set\n")
	}

	// import:
	if len(dimg.RawImport) != 1 {
		t.Fatalf("dimg.RawMount len must be 1")
	}

	if dimg.RawImport[0].RawDimg == nil {
		t.Fatalf("dimg.RawMount[0].RawDimg should be set\n")
	}

}
