// Package fakedockerd provides a minimal in-process Docker Engine API stub for
// tests that must exercise real werf code paths without touching a shared
// Docker daemon.
package fakedockerd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/docker"
)

// NewContext serves daemon over a local http server bound to the lifetime of
// the current spec and returns a context bound to a docker api client talking
// to it, so that the code under test never reaches the docker daemon of the
// host.
func NewContext(ctx context.Context, daemon *Daemon) context.Context {
	server := httptest.NewServer(daemon)
	ginkgo.DeferCleanup(server.Close)

	ginkgo.GinkgoT().Setenv("DOCKER_HOST", strings.Replace(server.URL, "http://", "tcp://", 1))
	ginkgo.GinkgoT().Setenv("DOCKER_TLS_VERIFY", "")
	ginkgo.GinkgoT().Setenv("DOCKER_CERT_PATH", "")

	dockerCtx, err := docker.NewContextWithStreams(ctx, io.Discard, io.Discard)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return dockerCtx
}

// Daemon serves the handful of Docker Engine API endpoints used by the werf
// purge code paths and records the mutations it was asked to perform.
//
// A volume exists for the daemon when one of Containers mounts it, and is in
// use as long as one of those containers has not been removed — the reference
// counting that makes `docker volume rm` fail with 409 on a real daemon.
// A fixture container with State.Running set cannot be removed without force,
// the same 409 the real daemon answers with.
// Containers created over the API (the cleanup service container of the docker
// backend) are served, but never recorded as removed: the recorded mutations
// are the ones performed on the fixtures.
type Daemon struct {
	Containers []container.InspectResponse

	// AnonymousVolumes are the volume names the daemon reports as anonymous, the
	// only ones a container removal with v=1 takes down with the container.
	AnonymousVolumes []string

	// ContainerInspectStatus, ContainerRemoveStatus and VolumeRemoveStatus make
	// the daemon answer an inspect/removal of the given container id or volume
	// name with an HTTP error status instead of performing it.
	ContainerInspectStatus map[string]int
	ContainerRemoveStatus  map[string]int
	VolumeRemoveStatus     map[string]int

	mu                sync.Mutex
	requests          []string
	removedContainers []string
	removedVolumes    []string
	createdContainers int
}

var (
	_ http.Handler = (*Daemon)(nil)

	apiVersionPrefix     = regexp.MustCompile(`^/v[0-9.]+`)
	containerInspectPath = regexp.MustCompile(`^/containers/([^/]+)/json$`)
	containerAttachPath  = regexp.MustCompile(`^/containers/([^/]+)/attach$`)
	containerStartPath   = regexp.MustCompile(`^/containers/([^/]+)/start$`)
	containerWaitPath    = regexp.MustCompile(`^/containers/([^/]+)/wait$`)
	containerRemovePath  = regexp.MustCompile(`^/containers/([^/]+)$`)
	volumeRemovePath     = regexp.MustCompile(`^/volumes/([^/]+)$`)
	imageInspectPath     = regexp.MustCompile(`^/images/(.+)/json$`)
)

// Requests returns the recorded API calls in the order they were served, as
// "METHOD /path[?query]" strings with the API version prefix stripped.
func (d *Daemon) Requests(_ context.Context) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.requests)
}

// RemovedContainers returns the ids of the fixture containers removed so far.
func (d *Daemon) RemovedContainers(_ context.Context) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.removedContainers)
}

// RemovedVolumes returns the names of the volumes removed so far.
func (d *Daemon) RemovedVolumes(_ context.Context) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.removedVolumes)
}

func (d *Daemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := apiVersionPrefix.ReplaceAllString(r.URL.Path, "")

	if path == "/_ping" {
		w.Header().Set("Api-Version", "1.51")
		w.Header().Set("Ostype", "linux")
		w.WriteHeader(http.StatusOK)
		return
	}

	d.mu.Lock()
	d.requests = append(d.requests, fmt.Sprintf("%s %s", r.Method, apiVersionPrefix.ReplaceAllString(r.URL.RequestURI(), "")))
	d.mu.Unlock()

	switch {
	case path == "/version":
		d.writeJSON(w, http.StatusOK, map[string]string{"ApiVersion": "1.51", "Version": "fake"})
	case path == "/containers/json":
		d.listContainers(w, r.URL.Query().Get("filters"))
	case path == "/images/json":
		d.writeJSON(w, http.StatusOK, []any{})
	case imageInspectPath.MatchString(path):
		d.writeError(w, http.StatusNotFound, "No such image")
	case path == "/containers/create":
		d.createContainer(w)
	case containerAttachPath.MatchString(path):
		hijackAndClose(w)
	case containerStartPath.MatchString(path):
		w.WriteHeader(http.StatusNoContent)
	case containerWaitPath.MatchString(path):
		d.writeJSON(w, http.StatusOK, container.WaitResponse{})
	case r.Method == http.MethodGet && containerInspectPath.MatchString(path):
		d.inspectContainer(w, containerInspectPath.FindStringSubmatch(path)[1])
	case r.Method == http.MethodDelete && containerRemovePath.MatchString(path):
		d.removeContainer(w, containerRemovePath.FindStringSubmatch(path)[1], r.URL.Query())
	case r.Method == http.MethodDelete && volumeRemovePath.MatchString(path):
		d.removeVolume(w, volumeRemovePath.FindStringSubmatch(path)[1], r.URL.Query())
	default:
		d.writeError(w, http.StatusNotImplemented, fmt.Sprintf("fake daemon: unexpected request %s %s", r.Method, path))
	}
}

func (d *Daemon) listContainers(w http.ResponseWriter, rawFilters string) {
	filters, err := parseContainerFilters(rawFilters)
	if err != nil {
		d.writeError(w, http.StatusNotImplemented, fmt.Sprintf("fake daemon: %s", err))
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	summaries := make([]container.Summary, 0, len(d.Containers))
	for _, c := range d.Containers {
		if slices.Contains(d.removedContainers, c.ID) {
			continue
		}

		matchesName := len(filters["name"]) == 0 || slices.ContainsFunc(filters["name"], func(f string) bool {
			return strings.Contains(c.Name, f)
		})
		if !matchesName {
			continue
		}

		matchesVolume := len(filters["volume"]) == 0 || slices.ContainsFunc(filters["volume"], func(f string) bool {
			return slices.ContainsFunc(c.Mounts, func(m container.MountPoint) bool {
				return m.Type == "volume" && m.Name == f
			})
		})
		if !matchesVolume {
			continue
		}

		summaries = append(summaries, container.Summary{ID: c.ID, Names: []string{c.Name}, Image: c.Image})
	}

	d.writeJSON(w, http.StatusOK, summaries)
}

// Values of the same filter key are OR-ed by the daemon, different keys are
// AND-ed; an unsupported key is rejected instead of silently widening the
// result.
func parseContainerFilters(rawFilters string) (map[string][]string, error) {
	if rawFilters == "" {
		return nil, nil
	}

	query, err := url.QueryUnescape(rawFilters)
	if err != nil {
		return nil, fmt.Errorf("unescape filters %q: %w", rawFilters, err)
	}

	var parsed map[string]map[string]bool
	if err := json.Unmarshal([]byte(query), &parsed); err != nil {
		return nil, fmt.Errorf("parse filters %q: %w", query, err)
	}

	filters := make(map[string][]string, len(parsed))
	for key, values := range parsed {
		if key != "name" && key != "volume" {
			return nil, fmt.Errorf("unsupported container filter %q", key)
		}
		for value, enabled := range values {
			if enabled {
				filters[key] = append(filters[key], value)
			}
		}
	}

	return filters, nil
}

func (d *Daemon) createContainer(w http.ResponseWriter) {
	d.mu.Lock()
	d.createdContainers++
	id := fmt.Sprintf("created-container-%d", d.createdContainers)
	d.mu.Unlock()

	d.writeJSON(w, http.StatusCreated, container.CreateResponse{ID: id})
}

func hijackAndClose(w http.ResponseWriter) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		panic(fmt.Sprintf("fake daemon: hijack attach connection: %s", err))
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n"); err != nil {
		panic(fmt.Sprintf("fake daemon: write attach response: %s", err))
	}
}

func (d *Daemon) inspectContainer(w http.ResponseWriter, ref string) {
	if status, ok := d.ContainerInspectStatus[ref]; ok {
		d.writeError(w, status, fmt.Sprintf("cannot inspect container %q", ref))
		return
	}

	d.mu.Lock()
	removed := slices.Clone(d.removedContainers)
	containers := d.Containers
	d.mu.Unlock()

	for _, c := range containers {
		if c.ID != ref && strings.TrimPrefix(c.Name, "/") != strings.TrimPrefix(ref, "/") {
			continue
		}
		if slices.Contains(removed, c.ID) {
			break
		}
		d.writeJSON(w, http.StatusOK, c)
		return
	}

	d.writeError(w, http.StatusNotFound, fmt.Sprintf("No such container: %s", ref))
}

func (d *Daemon) removeContainer(w http.ResponseWriter, id string, query url.Values) {
	if query.Get("link") == "1" {
		d.writeError(w, http.StatusNotImplemented, "fake daemon: link removal is not modeled")
		return
	}

	if status, ok := d.ContainerRemoveStatus[id]; ok {
		d.writeError(w, status, fmt.Sprintf("cannot remove container %q", id))
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	index := slices.IndexFunc(d.Containers, func(c container.InspectResponse) bool { return c.ID == id })
	if index == -1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	c := d.Containers[index]

	// The message is the one pkg/docker/errors.go recognizes as "container is running".
	if c.State != nil && c.State.Running && query.Get("force") != "1" {
		d.writeError(w, http.StatusConflict, fmt.Sprintf("cannot remove container %q: container is running: stop the container before removing or force remove", id))
		return
	}

	d.removedContainers = append(d.removedContainers, id)

	// v removes the anonymous volumes of the container only: named volumes outlive
	// their container, and a volume another container still mounts stays either way.
	if query.Get("v") == "1" {
		for _, m := range c.Mounts {
			if m.Type == "volume" && slices.Contains(d.AnonymousVolumes, m.Name) {
				_ = d.removeVolumeLocked(m.Name)
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

var (
	errVolumeInUse  = errors.New("volume is in use")
	errNoSuchVolume = errors.New("no such volume")
)

func (d *Daemon) removeVolume(w http.ResponseWriter, name string, query url.Values) {
	if status, ok := d.VolumeRemoveStatus[name]; ok {
		d.writeError(w, status, fmt.Sprintf("cannot remove volume %q", name))
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.removeVolumeLocked(name)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, errVolumeInUse):
		d.writeError(w, http.StatusConflict, err.Error())
	case query.Get("force") == "1":
		// force makes the daemon ignore a volume that is already gone, nothing else.
		w.WriteHeader(http.StatusNoContent)
	default:
		d.writeError(w, http.StatusNotFound, err.Error())
	}
}

// removeVolumeLocked removes the volume if it exists and no container that
// mounts it is left. d.mu must be held.
func (d *Daemon) removeVolumeLocked(name string) error {
	var exists bool
	for _, c := range d.Containers {
		if !slices.ContainsFunc(c.Mounts, func(m container.MountPoint) bool { return m.Type == "volume" && m.Name == name }) {
			continue
		}

		exists = true
		if !slices.Contains(d.removedContainers, c.ID) {
			return fmt.Errorf("remove %s: %w - [%s]", name, errVolumeInUse, c.ID)
		}
	}

	if !exists || slices.Contains(d.removedVolumes, name) {
		return fmt.Errorf("get %s: %w", name, errNoSuchVolume)
	}

	d.removedVolumes = append(d.removedVolumes, name)

	return nil
}

func (d *Daemon) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		panic(fmt.Sprintf("fake daemon: encode response: %s", err))
	}
}

func (d *Daemon) writeError(w http.ResponseWriter, status int, message string) {
	d.writeJSON(w, status, map[string]string{"message": message})
}
