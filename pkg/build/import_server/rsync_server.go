package import_server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/stapel"
)

const rsyncServerPort = "873"

type RsyncServer struct {
	IPAddress              string
	Port                   string
	DockerContainerName    string
	DockerImageName        string
	AuthUser, AuthPassword string
}

func RunRsyncServer(ctx context.Context, dockerImageName, tmpDir string) (*RsyncServer, error) {
	logboek.Context(ctx).Debug().LogF("RunRsyncServer for docker image %q\n", dockerImageName)

	srv := &RsyncServer{
		Port:                rsyncServerPort,
		DockerContainerName: fmt.Sprintf("%s%s", image.ImportServerContainerNamePrefix, uuid.New().String()),
		AuthUser:            fmt.Sprintf("werf-%s", generateSecureRandomString(4)),
		AuthPassword:        generateSecureRandomString(16),
	}

	stapelContainerName, err := stapel.GetOrCreateContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get or create stapel container: %w", err)
	}

	secretsFilePath := path.Join(tmpDir, "rsyncd.secrets")
	if err := ioutil.WriteFile(secretsFilePath, []byte(fmt.Sprintf("%s:%s\n", srv.AuthUser, srv.AuthPassword)), 0o644); err != nil {
		return nil, fmt.Errorf("unable to write %s: %w", secretsFilePath, err)
	}

	rsyncConfPath := path.Join(tmpDir, "rsyncd.conf")
	if err := ioutil.WriteFile(rsyncConfPath, []byte(fmt.Sprintf(`pid file = /.werf/rsyncd.pid
lock file = /.werf/rsyncd.lock
log file = /.werf/rsyncd.log
uid = 0
gid = 0
port = %s

[import]
path = /
comment = Image files to import
read only = true
timeout = 300
auth users = %s
secrets file = /.werf/rsyncd.secrets
strict modes = false
`, rsyncServerPort, srv.AuthUser)), 0o644); err != nil {
		return nil, fmt.Errorf("unable to write %s: %w", rsyncConfPath, err)
	}

	runArgs := []string{
		"--detach",
		"--rm",
		"--user=0:0",
		"--workdir=/",
		fmt.Sprintf("--name=%s", srv.DockerContainerName),
		fmt.Sprintf("--volumes-from=%s", stapelContainerName),
		fmt.Sprintf("--volume=%s:/.werf/rsyncd.conf", rsyncConfPath),
		fmt.Sprintf("--volume=%s:/.werf/rsyncd.secrets", secretsFilePath),
		fmt.Sprintf("--expose=%s", rsyncServerPort),
		fmt.Sprintf("--entrypoint=%s", stapel.RsyncBinPath()),
		dockerImageName,
		"--daemon",
		"--no-detach",
		"--config=/.werf/rsyncd.conf",
	}
	logboek.Context(ctx).Debug().LogF("Run rsync server command: %q\n", fmt.Sprintf("docker run %s", strings.Join(runArgs, " ")))
	if output, err := docker.CliRun_RecordedOutput(ctx, runArgs...); err != nil {
		logboek.Context(ctx).Error().LogF("Unable to run rsync server command: %q\n", fmt.Sprintf("docker run %s", strings.Join(runArgs, " ")))
		logboek.Context(ctx).Error().LogF("%s", output)
		return nil, err
	}

	logboek.Context(ctx).Debug().LogF("Inspect container %s\n", srv.DockerContainerName)

	if inspect, err := docker.ContainerInspect(ctx, srv.DockerContainerName); err != nil {
		return nil, fmt.Errorf("unable to inspect import server container %s: %w", srv.DockerContainerName, err)
	} else {
		if inspect.NetworkSettings == nil {
			return nil, fmt.Errorf("unable to get import server container %s ip address: no network settings available in inspect", srv.DockerContainerName)
		}
		srv.IPAddress = inspect.NetworkSettings.Networks["bridge"].IPAddress
	}

	return srv, nil
}

func (srv *RsyncServer) Shutdown(ctx context.Context) error {
	if output, err := docker.CliRm_RecordedOutput(ctx, "--force", srv.DockerContainerName); err != nil {
		logboek.Context(ctx).Error().LogF("%s", output)
		return fmt.Errorf("unable to remove container %s: %w", srv.DockerContainerName, err)
	}
	return nil
}

func (srv *RsyncServer) GetCopyCommand(ctx context.Context, importConfig *config.Import) string {
	var args []string

	rsyncImportPathSpec := fmt.Sprintf("rsync://%s@%s:%s/import/%s", srv.AuthUser, srv.IPAddress, srv.Port, importConfig.Add)
	rsyncStatImportPathCommand := fmt.Sprintf("RSYNC_PASSWORD='%s' %s -L %s", srv.AuthPassword, stapel.RsyncBinPath(), rsyncImportPathSpec)

	// save stat output to variable
	args = append(args, fmt.Sprintf("statOutput=$(%s)", rsyncStatImportPathCommand))
	// check command exit code from last subshell
	args = append(args, "[ $? -eq 0 ]")
	// unset old value of IMPORT_PATH_TRAILING_SLASH_OPTIONAL variable from other copy commands
	args = append(args, "unset IMPORT_PATH_TRAILING_SLASH_OPTIONAL")
	// set fileTypeField
	args = append(args, fmt.Sprintf("fileTypeField=$(echo $statOutput | %s -c1)", stapel.HeadBinPath()))
	// check command exit code from last subshell
	args = append(args, "[ $? -eq 0 ]")
	// set optional trailing slash when importing directory so that rsync will automatically
	// merge already existing directory in the target image
	args = append(args, "if [ $fileTypeField = d ] ; then IMPORT_PATH_TRAILING_SLASH_OPTIONAL=/ ; fi")
	// create a parent directory where target file/directory will reside
	args = append(args, fmt.Sprintf("%s -p %s", stapel.MkdirBinPath(), path.Dir(importConfig.To)))

	var rsyncChownOption string
	if importConfig.Owner != "" || importConfig.Group != "" {
		rsyncChownOption = fmt.Sprintf("--chown=%s:%s", importConfig.Owner, importConfig.Group)
	}
	rsyncCommand := fmt.Sprintf("RSYNC_PASSWORD='%s' %s --archive --links --inplace --xattrs %s", srv.AuthPassword, stapel.RsyncBinPath(), rsyncChownOption)
	rsyncCommand += PrepareRsyncFilters(importConfig.Add, importConfig.IncludePaths, importConfig.ExcludePaths)

	rsyncCommand += fmt.Sprintf(" %s$IMPORT_PATH_TRAILING_SLASH_OPTIONAL %s", rsyncImportPathSpec, importConfig.To)
	// run rsync itself
	args = append(args, rsyncCommand)

	command := strings.Join(args, " && ")

	logboek.Context(ctx).Debug().LogF("Rsync server copy commands for import: artifact=%q image=%q add=%s to=%s includePaths=%v excludePaths=%v: %q\n", importConfig.ArtifactName, importConfig.ImageName, importConfig.Add, importConfig.To, importConfig.IncludePaths, importConfig.ExcludePaths, command)

	return command
}

func PrepareRsyncFilters(add string, includePaths, excludePaths []string) string {
	rsyncCommand := ""
	if len(includePaths) != 0 {
		// First, apply exclude filters to the specified paths.
		rsyncCommand += PrepareRsyncExcludeFiltersForGlobs(add, excludePaths)
		// Then include only the paths that are listed in include_paths.
		rsyncCommand += PrepareRsyncIncludeFiltersForGlobs(add, includePaths)
	} else if len(excludePaths) != 0 {
		// When include_paths is empty, simply apply exclude filters.
		rsyncCommand += PrepareRsyncExcludeFiltersForGlobs(add, excludePaths)
	}
	return rsyncCommand
}

// PrepareRsyncExcludeFiltersForGlobs builds rsync --filter rules that exclude
// paths matching given globs under the specified base path (add).
// It uses globToRsyncFilterPaths with finalOnly=true to generate the
// minimal set of patterns needed to prevent rsync from descending into
// excluded directories and files.
//
// For each excludeGlob in excludeGlobs, it generates rules like:
//
//	--filter='-/ base/excludeGlobPrefix...'
func PrepareRsyncExcludeFiltersForGlobs(add string, excludeGlobs []string) string {
	if len(excludeGlobs) == 0 {
		return ""
	}

	paths := map[string]struct{}{}
	for _, p := range excludeGlobs {
		targetPath := path.Join(add, p)
		for _, pathPart := range globToRsyncFilterPaths(targetPath, true) {
			paths[pathPart] = struct{}{}
		}
	}
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b string
	for _, p := range keys {
		b += fmt.Sprintf(" --filter='-/ %s'", p)
	}
	return b
}

// PrepareRsyncIncludeFiltersForGlobs builds rsync --filter rules that include
// only the specified includeGlobs under the given base path (add). It uses
// globToRsyncFilterPaths with finalOnly=false to ensure that all parent
// directories for included paths are traversable, and then adds final
// rules for the glob itself and its recursive contents.
//
// For each includeGlob in includeGlobs, it generates rules like:
//
//	--filter='+/ base/...prefixes.../'
//	--filter='+/ base/includeGlob'
//	--filter='+/ base/includeGlob/**'
//
// At the end, it adds a catch-all exclude:
//
//	--filter='-/ base/**'
func PrepareRsyncIncludeFiltersForGlobs(add string, includeGlobs []string) string {
	if len(includeGlobs) == 0 {
		return ""
	}

	paths := map[string]struct{}{}
	for _, p := range includeGlobs {
		targetPath := path.Join(add, p)

		// Allow all path prefixes for this glob.
		for _, pathPart := range globToRsyncFilterPaths(targetPath, false) {
			paths[pathPart] = struct{}{}
		}

		// We do not know in advance whether it is a file or a directory — add both variants.
		paths[targetPath] = struct{}{}
		paths[path.Join(targetPath, "**")] = struct{}{}
	}
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b string
	for _, k := range keys {
		b += fmt.Sprintf(" --filter='+/ %s'", k)
	}
	// Everything that did not match any include is excluded.
	b += fmt.Sprintf(" --filter='-/ %s'", path.Join(add, "**"))
	return b
}

// globToRsyncFilterPaths builds rsync filter path components for a glob.
// Behavior:
// - Directories (non-final segments) end with "/" unless finalOnly == true.
// - Final segment (file or pattern) has no trailing "/".
// - "**" expands into two branches: keep ("**") and skip (matches 0 directories).
// - When kept and not final && !finalOnly -> adds "**/" as a directory pattern.
// Empty or all-slash input returns nil.
func globToRsyncFilterPaths(glob string, finalSegmentOnly bool) []string {
	glob = strings.Trim(glob, "/")
	if glob == "" {
		return nil
	}

	segments := strings.Split(glob, "/")
	lastIdx := len(segments) - 1

	type void struct{}
	set := func(m map[string]void, v string) {
		if v != "" {
			m[v] = void{}
		}
	}

	current := map[string]void{"": {}}
	results := map[string]void{}

	join := func(prefix, seg string) string {
		if prefix == "" {
			return seg
		}
		return prefix + "/" + seg
	}

	for i, seg := range segments {
		next := map[string]void{}
		isLast := i == lastIdx

		if seg == "**" {
			for prefix := range current {
				// Branch: keep "**"
				keep := join(prefix, "**")
				set(next, keep)

				// "**" as a directory (recursive) when not final and we collect dirs
				if !isLast && !finalSegmentOnly {
					set(results, keep+"/")
				}

				// Branch: skip "**" (0 directories matched)
				set(next, prefix)

				// Nothing added to results for skip branch (prefix stays as-is).
			}
		} else {
			for prefix := range current {
				full := join(prefix, seg)
				set(next, full)

				if isLast {
					// Final pattern/file
					set(results, full)
				} else if !finalSegmentOnly {
					// Intermediate directory
					set(results, full+"/")
				}
			}
		}

		current = next
	}

	out := make([]string, 0, len(results))
	for v := range results {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func generateSecureRandomString(length int) string {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(fmt.Sprintf("cannot generate secure random string: %s", err))
	}
	return hex.EncodeToString(randomBytes)
}
