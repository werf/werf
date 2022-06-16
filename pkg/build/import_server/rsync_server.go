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
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/stapel"
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
		DockerContainerName: fmt.Sprintf("import-server-%s", uuid.New().String()),
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
		srv.IPAddress = inspect.NetworkSettings.IPAddress
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
	rsyncCommand := fmt.Sprintf("RSYNC_PASSWORD='%s' %s --archive --links --inplace %s", srv.AuthPassword, stapel.RsyncBinPath(), rsyncChownOption)

	if len(importConfig.IncludePaths) != 0 {
		/**
				Если указали include_paths — это означает, что надо копировать
				только указанные пути. Поэтому exclude_paths в приоритете, т.к. в данном режиме
		        exclude_paths может относится только к путям, указанным в include_paths.
		        При этом случай, когда в include_paths указали более специальный путь, чем в exclude_paths,
		        будет обрабатываться в пользу exclude, этот путь не скопируется.
		*/
		for _, p := range importConfig.ExcludePaths {
			rsyncCommand += fmt.Sprintf(" --filter='-/ %s'", path.Join(importConfig.Add, p))
		}

		for _, p := range importConfig.IncludePaths {
			targetPath := path.Join(importConfig.Add, p)

			// Генерируем разрешающее правило для каждого элемента пути
			for _, pathPart := range descentPath(targetPath) {
				rsyncCommand += fmt.Sprintf(" --filter='+/ %s'", pathPart)
			}

			/**
					На данный момент не знаем директорию или файл имел в виду пользователь,
			        поэтому подставляем фильтры для обоих возможных случаев.

					Автоматом подставляем паттерн ** для включения файлов, содержащихся в
			        директории, которую пользователь указал в include_paths.
			*/
			rsyncCommand += fmt.Sprintf(" --filter='+/ %s'", targetPath)
			rsyncCommand += fmt.Sprintf(" --filter='+/ %s'", path.Join(targetPath, "**"))
		}

		// Все что не подошло по include — исключается
		rsyncCommand += fmt.Sprintf(" --filter='-/ %s'", path.Join(importConfig.Add, "**"))
	} else {
		for _, p := range importConfig.ExcludePaths {
			rsyncCommand += fmt.Sprintf(" --filter='-/ %s'", path.Join(importConfig.Add, p))
		}
	}

	rsyncCommand += fmt.Sprintf(" %s$IMPORT_PATH_TRAILING_SLASH_OPTIONAL %s", rsyncImportPathSpec, importConfig.To)
	// run rsync itself
	args = append(args, rsyncCommand)

	command := strings.Join(args, " && ")

	logboek.Context(ctx).Debug().LogF("Rsync server copy commands for import: artifact=%q image=%q add=%s to=%s includePaths=%v excludePaths=%v: %q\n", importConfig.ArtifactName, importConfig.ImageName, importConfig.Add, importConfig.To, importConfig.IncludePaths, importConfig.ExcludePaths, command)

	return command
}

func descentPath(filePath string) []string {
	var parts []string

	part := filePath
	for {
		parts = append(parts, part)
		part = path.Dir(part)

		if part == path.Dir(part) {
			break
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(parts)))

	return parts
}

func generateSecureRandomString(length int) string {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(fmt.Sprintf("cannot generate secure random string: %s", err))
	}
	return hex.EncodeToString(randomBytes)
}
