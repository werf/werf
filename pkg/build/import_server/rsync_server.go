package import_server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/flant/werf/pkg/docker"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/stapel"
)

const rsyncServerPort = "873"

type RsyncServer struct {
	IPAddress              string
	Port                   string
	DockerContainerName    string
	DockerImageName        string
	AuthUser, AuthPassword string
}

func RunRsyncServer(dockerImageName string, tmpDir string) (*RsyncServer, error) {
	logboek.Debug.LogF("RunRsyncServer for docker image %q\n", dockerImageName)

	srv := &RsyncServer{
		Port:                rsyncServerPort,
		DockerContainerName: fmt.Sprintf("import-server-%s", uuid.New().String()),
		AuthUser:            fmt.Sprintf("werf-%s", generateSecureRandomString(4)),
		AuthPassword:        generateSecureRandomString(16),
	}

	stapelContainerName, err := stapel.GetOrCreateContainer()
	if err != nil {
		return nil, err
	}

	secretsFilePath := path.Join(tmpDir, "rsyncd.secrets")
	if err := ioutil.WriteFile(secretsFilePath, []byte(fmt.Sprintf("%s:%s\n", srv.AuthUser, srv.AuthPassword)), 0644); err != nil {
		return nil, fmt.Errorf("unable to write %s: %s", secretsFilePath, err)
	}

	rsyncConfPath := path.Join(tmpDir, "rsyncd.conf")
	if err := ioutil.WriteFile(rsyncConfPath, []byte(fmt.Sprintf(`pid file = /.werf/rsyncd.pid
lock file = /.werf/rsyncd.lock
log file = /.werf/rsyncd.log
uid = root
port = %s

[import]
path = /
comment = Image files to import
read only = true
timeout = 300
auth users = %s
secrets file = /.werf/rsyncd.secrets
strict modes = false
`, rsyncServerPort, srv.AuthUser)), 0644); err != nil {
		return nil, fmt.Errorf("unable to write %s: %s", rsyncConfPath, err)
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
	logboek.Debug.LogF("Run rsync server command: %q\n", fmt.Sprintf("docker run %s", strings.Join(runArgs, " ")))
	if output, err := docker.CliRun_RecordedOutput(runArgs...); err != nil {
		logboek.LogErrorF("%s", output)
		return nil, err
	}

	logboek.Debug.LogF("Inspect container %s\n", srv.DockerContainerName)

	if inspect, err := docker.ContainerInspect(srv.DockerContainerName); err != nil {
		return nil, fmt.Errorf("unable to inspect import server container %s: %s", srv.DockerContainerName, err)
	} else {
		if inspect.NetworkSettings == nil {
			return nil, fmt.Errorf("unable to get import server container %s ip address: no network settings available in inspect")
		}
		srv.IPAddress = inspect.NetworkSettings.IPAddress
	}

	return srv, nil
}

func (srv *RsyncServer) Shutdown() error {
	if output, err := docker.CliRm_RecordedOutput("--force", srv.DockerContainerName); err != nil {
		logboek.LogErrorF("%s", output)
		return fmt.Errorf("unable to remove container %s: %s", srv.DockerContainerName, err)
	}
	return nil
}

func (srv *RsyncServer) GetCopyCommand(importConfig *config.Import) string {
	var args []string

	mkdirBin := stapel.MkdirBinPath()
	mkdirPath := path.Dir(importConfig.To)
	mkdirCommand := fmt.Sprintf("%s -p %s", mkdirBin, mkdirPath)

	rsyncBin := stapel.RsyncBinPath()
	var rsyncChownOption string
	if importConfig.Owner != "" || importConfig.Group != "" {
		rsyncChownOption = fmt.Sprintf("--chown=%s:%s", importConfig.Owner, importConfig.Group)
	}
	rsyncCommand := fmt.Sprintf("RSYNC_PASSWORD='%s' %s --archive --links --inplace %s", srv.AuthPassword, rsyncBin, rsyncChownOption)

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

	rsyncCommand += fmt.Sprintf(" rsync://%s@%s:%s/import/%s %s", srv.AuthUser, srv.IPAddress, srv.Port, importConfig.Add, path.Dir(importConfig.To))

	args = append(args, mkdirCommand, rsyncCommand)

	addBase := path.Base(importConfig.Add)
	toBase := path.Base(importConfig.To)
	if addBase != toBase && toBase != "" {
		args = append(args, fmt.Sprintf("%s %s %s", stapel.MvBinPath(), path.Join(path.Dir(importConfig.To), addBase), path.Join(path.Dir(importConfig.To), toBase)))
	}

	command := strings.Join(args, " && ")

	logboek.Debug.LogF("Rsync server copy commands for import: artifact=%q image=%q add=%s to=%s includePaths=%v excludePaths=%v: %q\n", importConfig.ArtifactName, importConfig.ImageName, importConfig.Add, importConfig.To, importConfig.IncludePaths, importConfig.ExcludePaths, command)

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

	sort.Sort(sort.Reverse(sort.StringSlice(parts[:])))

	return parts
}

func generateSecureRandomString(lenght int) string {
	randomBytes := make([]byte, lenght)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(fmt.Sprintf("cannot generate secure random string: %s", err))
	}
	return hex.EncodeToString(randomBytes)
}
