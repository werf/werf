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

func Test_ParseDimgs_Git(t *testing.T) {
	dimgs, err := ParseDimgs("testdata/git.yaml")
	if err != nil {
		t.Fatal(err)
	}

	dimgGitLocals := dimgs[0].Git.Local
	for _, git := range dimgGitLocals {
		git.Raw = nil
		git.GitLocalExport.Raw = nil
		git.GitExport.Raw = nil
		git.GitExportBase.Raw = nil
		git.ExportBase.Raw = nil
		git.StageDependencies.Raw = nil
	}
	dimgGitRemotes := dimgs[0].Git.Remote
	for _, git := range dimgGitRemotes {
		git.Raw = nil
		git.GitRemoteExport.Raw = nil
		git.GitLocalExport.Raw = nil
		git.GitExport.Raw = nil
		git.GitExportBase.Raw = nil
		git.ExportBase.Raw = nil
		git.StageDependencies.Raw = nil
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

	artifactGitLocals := dimgs[0].Import[0].ArtifactDimg.Git.Local
	for _, git := range artifactGitLocals {
		git.Raw = nil
		git.GitLocalExport.Raw = nil
		git.GitExport.Raw = nil
		git.GitExportBase.Raw = nil
		git.ExportBase.Raw = nil
		git.StageDependencies.Raw = nil
	}
	artifactGitRemotes := dimgs[0].Import[0].ArtifactDimg.Git.Remote
	for _, git := range artifactGitRemotes {
		git.Raw = nil
		git.GitRemoteExport.Raw = nil
		git.GitLocalExport.Raw = nil
		git.GitExport.Raw = nil
		git.GitExportBase.Raw = nil
		git.ExportBase.Raw = nil
		git.StageDependencies.Raw = nil
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
	shellDimg.Raw = nil
	shellDimg.ShellBase.Raw = nil
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

	shellArtifact := dimgs[0].Import[0].ArtifactDimg.Shell
	shellArtifact.Raw = nil
	shellArtifact.ShellBase.Raw = nil
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
	ansibleDimg.Raw = nil
	ansibleDimg.BeforeInstall[0].Raw = nil
	ansibleDimg.Install[0].Raw = nil
	ansibleDimg.BeforeSetup[0].Raw = nil
	ansibleDimg.Setup[0].Raw = nil
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

	ansibleArtifact := dimgs[0].Import[0].ArtifactDimg.Ansible
	ansibleArtifact.Raw = nil
	ansibleArtifact.Raw = nil
	ansibleArtifact.BeforeInstall[0].Raw = nil
	ansibleArtifact.Install[0].Raw = nil
	ansibleArtifact.BeforeSetup[0].Raw = nil
	ansibleArtifact.Setup[0].Raw = nil
	ansibleArtifact.BuildArtifact[0].Raw = nil
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
	docker.Raw = nil
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
		mount.Raw = nil
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
