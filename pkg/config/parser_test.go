package config

import (
	"reflect"
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

	docs := []*doc{
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

	if dimg.doc == nil {
		t.Fatalf("dimg.Doc should be set\n")
	}

	// git:
	if len(dimg.RawGit) != 1 {
		t.Fatalf("dimg.rawGit len must be 1")
	}

	if dimg.RawGit[0].rawDimg == nil {
		t.Fatalf("dimg.rawGit[0].rawDimg should be set\n")
	}

	if dimg.RawGit[0].RawStageDependencies == nil {
		t.Fatalf("dimg.rawGit[0].rawStageDependencies must be set")
	}

	if dimg.RawGit[0].RawStageDependencies.rawGit == nil {
		t.Fatalf("dimg.rawGit[0].rawStageDependencies.rawShell should be set\n")
	}

	// shell:
	if dimg.RawShell.rawDimg == nil {
		t.Fatalf("dimg.rawShell.rawDimg should be set\n")
	}

	// mount:
	if len(dimg.RawMount) != 1 {
		t.Fatalf("dimg.rawMount len must be 1")
	}

	if dimg.RawMount[0].rawDimg == nil {
		t.Fatalf("dimg.rawMount[0].rawDimg should be set\n")
	}

	// docker:
	if dimg.RawDocker.rawDimg == nil {
		t.Fatalf("dimg.rawDocker.rawDimg should be set\n")
	}

	// import:
	if len(dimg.RawImport) != 1 {
		t.Fatalf("dimg.rawMount len must be 1")
	}

	if dimg.RawImport[0].rawDimg == nil {
		t.Fatalf("dimg.rawMount[0].rawDimg should be set\n")
	}

}

func Test_ParseDimgs_Git(t *testing.T) {
	dimgs, err := ParseDimgs("testdata/git.yaml")
	if err != nil {
		t.Fatal(err)
	}

	dimgGitLocals := dimgs[0].Git.Local
	for _, git := range dimgGitLocals {
		git.raw = nil
		git.GitLocalExport.raw = nil
		git.GitExport.raw = nil
		git.GitExportBase.raw = nil
		git.ExportBase.raw = nil
		git.StageDependencies.raw = nil
	}
	dimgGitRemotes := dimgs[0].Git.Remote
	for _, git := range dimgGitRemotes {
		git.raw = nil
		git.GitRemoteExport.raw = nil
		git.GitLocalExport.raw = nil
		git.GitExport.raw = nil
		git.GitExportBase.raw = nil
		git.ExportBase.raw = nil
		git.StageDependencies.raw = nil
	}
	expectedDimgGitLocals := []*GitLocal{
		{
			GitLocalExport: &GitLocalExport{
				GitExportBase: &GitExportBase{
					GitExport: &GitExport{
						ExportBase: &ExportBase{
							Add:          "/sub-folder",
							To:           "/local_git",
							IncludePaths: []string{"sub-sub1-folder"},
							ExcludePaths: []string{"sub-sub2-folder"},
							Owner:        "owner",
							Group:        "group",
						},
					},
					StageDependencies: &StageDependencies{
						Install:       []string{"**/*"},
						Setup:         []string{"file"},
						BeforeSetup:   []string{"sub-sub3-folder"},
						BuildArtifact: []string{},
					},
				},
			},
			As: "local_git",
		},
	}

	expectedDimgGitRemotes := []*GitRemote{
		{
			Url:  "git@github.com:company/project.git",
			Name: "company/project",
			GitRemoteExport: &GitRemoteExport{
				GitLocalExport: &GitLocalExport{
					GitExportBase: &GitExportBase{
						GitExport: &GitExport{
							ExportBase: &ExportBase{
								Add:          "/sub-folder",
								To:           "/remote_git",
								IncludePaths: []string{"sub-sub1-folder"},
								ExcludePaths: []string{"sub-sub2-folder"},
								Owner:        "owner",
								Group:        "group",
							},
						},
						StageDependencies: &StageDependencies{
							Install:       []string{"**/*"},
							Setup:         []string{"file"},
							BeforeSetup:   []string{"sub-sub3-folder"},
							BuildArtifact: []string{},
						},
					},
				},
			},
			As: "remote_git",
		},
	}

	if !reflect.DeepEqual(dimgGitLocals, expectedDimgGitLocals) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", dimgGitLocals, expectedDimgGitLocals)
	}

	if !reflect.DeepEqual(dimgGitRemotes, expectedDimgGitRemotes) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", dimgGitRemotes, expectedDimgGitRemotes)
	}

	artifactGitLocals := dimgs[0].Import[0].artifactDimg.Git.Local
	for _, git := range artifactGitLocals {
		git.raw = nil
		git.GitLocalExport.raw = nil
		git.GitExport.raw = nil
		git.GitExportBase.raw = nil
		git.ExportBase.raw = nil
		git.StageDependencies.raw = nil
	}
	artifactGitRemotes := dimgs[0].Import[0].artifactDimg.Git.Remote
	for _, git := range artifactGitRemotes {
		git.raw = nil
		git.GitRemoteExport.raw = nil
		git.GitLocalExport.raw = nil
		git.GitExport.raw = nil
		git.GitExportBase.raw = nil
		git.ExportBase.raw = nil
		git.StageDependencies.raw = nil
	}
	expectedArtifactGitLocals := expectedDimgGitLocals
	expectedArtifactGitLocals[0].StageDependencies.BuildArtifact = []string{"*.php"}

	expectedArtifactGitRemotes := expectedDimgGitRemotes
	expectedArtifactGitRemotes[0].StageDependencies.BuildArtifact = []string{"*.php"}

	if !reflect.DeepEqual(artifactGitLocals, expectedArtifactGitLocals) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", artifactGitLocals, expectedArtifactGitLocals)
	}

	if !reflect.DeepEqual(artifactGitRemotes, expectedArtifactGitRemotes) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", artifactGitRemotes, expectedArtifactGitRemotes)
	}
}

func Test_ParseDimgs_Shell(t *testing.T) {
	dimgs, err := ParseDimgs("testdata/shell.yaml")
	if err != nil {
		t.Fatal(err)
	}

	shellDimg := dimgs[0].Shell
	shellDimg.raw = nil
	shellDimg.ShellBase.raw = nil
	expectedShellDimg := &ShellDimg{
		ShellBase: &ShellBase{
			BeforeInstall:             []string{"cat \"beforeInstall\""},
			Install:                   []string{"cat \"install\""},
			BeforeSetup:               []string{"cat \"beforeSetup\""},
			Setup:                     []string{"cat \"setup\""},
			CacheVersion:              "cacheVersion",
			BeforeInstallCacheVersion: "beforeInstallCacheVersion",
			InstallCacheVersion:       "installCacheVersion",
			BeforeSetupCacheVersion:   "beforeSetupCacheVersion",
			SetupCacheVersion:         "setupCacheVersion",
		},
	}

	if !reflect.DeepEqual(shellDimg, expectedShellDimg) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", shellDimg, expectedShellDimg)
	}

	shellArtifact := dimgs[0].Import[0].artifactDimg.Shell
	shellArtifact.raw = nil
	shellArtifact.ShellBase.raw = nil
	expectedShellArtifact := &ShellArtifact{
		ShellDimg:                 expectedShellDimg,
		BuildArtifact:             []string{"cat \"buildArtifact\""},
		BuildArtifactCacheVersion: "buildArtifactCacheVersion",
	}

	if !reflect.DeepEqual(shellArtifact, expectedShellArtifact) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", shellArtifact, expectedShellArtifact)
	}
}

func Test_ParseDimgs_Ansible(t *testing.T) {
	dimgs, err := ParseDimgs("testdata/ansible.yaml")
	if err != nil {
		t.Fatal(err)
	}

	ansibleDimg := dimgs[0].Ansible
	ansibleDimg.raw = nil
	ansibleDimg.BeforeInstall[0].raw = nil
	ansibleDimg.Install[0].raw = nil
	ansibleDimg.BeforeSetup[0].raw = nil
	ansibleDimg.Setup[0].raw = nil
	expectedDimgAnsible := &Ansible{
		BeforeInstall: []*AnsibleTask{
			{Config: map[string]interface{}{"debug": map[interface{}]interface{}{"msg": "beforeInstall"}}},
		},
		Install: []*AnsibleTask{
			{Config: map[string]interface{}{"debug": map[interface{}]interface{}{"msg": "install"}}},
		},
		BeforeSetup: []*AnsibleTask{
			{Config: map[string]interface{}{"debug": map[interface{}]interface{}{"msg": "beforeSetup"}}},
		},
		Setup: []*AnsibleTask{
			{Config: map[string]interface{}{"debug": map[interface{}]interface{}{"msg": "setup"}}},
		},
		CacheVersion:              "cacheVersion",
		BeforeInstallCacheVersion: "beforeInstallCacheVersion",
		InstallCacheVersion:       "installCacheVersion",
		BeforeSetupCacheVersion:   "beforeSetupCacheVersion",
		SetupCacheVersion:         "setupCacheVersion",
	}

	if !reflect.DeepEqual(ansibleDimg, expectedDimgAnsible) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", ansibleDimg, expectedDimgAnsible)
	}

	ansibleArtifact := dimgs[0].Import[0].artifactDimg.Ansible
	ansibleArtifact.raw = nil
	ansibleArtifact.raw = nil
	ansibleArtifact.BeforeInstall[0].raw = nil
	ansibleArtifact.Install[0].raw = nil
	ansibleArtifact.BeforeSetup[0].raw = nil
	ansibleArtifact.Setup[0].raw = nil
	ansibleArtifact.BuildArtifact[0].raw = nil
	expectedAnsibleArtifact := expectedDimgAnsible
	expectedAnsibleArtifact.BuildArtifact = []*AnsibleTask{{Config: map[string]interface{}{"debug": map[interface{}]interface{}{"msg": "buildArtifact"}}}}
	expectedAnsibleArtifact.BuildArtifactCacheVersion = "buildArtifactCacheVersion"

	if !reflect.DeepEqual(ansibleArtifact, expectedAnsibleArtifact) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", ansibleArtifact, expectedAnsibleArtifact)
	}
}

func Test_ParseDimgs_Docker(t *testing.T) {
	dimgs, err := ParseDimgs("testdata/docker.yaml")
	if err != nil {
		t.Fatal(err)
	}

	docker := dimgs[0].Docker
	docker.raw = nil
	expectedDocker := &Docker{
		Volume:     []string{"/data"},
		Expose:     []string{"80/tcp"},
		Env:        map[string]string{"myName": "John Doe"},
		Label:      map[string]string{"com.example.vendor": "ACME Incorporated"},
		Cmd:        []string{"executable2", "param1", "param2"},
		Onbuild:    []string{"RUN /usr/local/bin/python-build --dir /app/src"},
		Workdir:    "folder",
		User:       "user",
		Entrypoint: []string{"executable1", "param1", "param2"},
	}

	if !reflect.DeepEqual(docker, expectedDocker) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", docker, expectedDocker)
	}
}

func Test_ParseDimgs_Mount(t *testing.T) {
	dimgs, err := ParseDimgs("testdata/mount.yaml")
	if err != nil {
		t.Fatal(err)
	}

	mounts := dimgs[0].Mount
	for _, mount := range dimgs[0].Mount {
		mount.raw = nil
	}
	expectedMounts := []*Mount{
		{To: "/folder1", From: "", Type: "build_dir"},
		{To: "/folder2", From: "", Type: "tmp_dir"},
		{To: "/folder3", From: "/from_path", Type: "custom_dir"},
	}

	if !reflect.DeepEqual(mounts, expectedMounts) {
		t.Errorf("\n[EXPECTED]: %#v\n[GOT]: %#v", mounts, expectedMounts)
	}
}
