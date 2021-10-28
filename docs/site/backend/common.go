package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ChannelType struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ReleaseType struct {
	Group    string        `yaml:"name"`
	Channels []ChannelType `yaml:"channels"`
}

type ReleasesStatusType struct {
	Releases []ReleaseType `yaml:"groups"`
}

type ApiStatusResponseType struct {
	Status         string        `json:"status"`
	Msg            string        `json:"msg"`
	RootVersion    string        `json:"rootVersion"`
	RootVersionURL string        `json:"rootVersionURL"`
	Multiwerf      []ReleaseType `json:"multiwerf"`
}

type versionMenuType struct {
	VersionItems           []versionMenuItems
	HTMLContent            string
	CurrentGroup           string
	CurrentChannel         string
	CurrentVersion         string
	AbsoluteVersion        string // Contains explicit version, used for getting git link to source file
	CurrentVersionURL      string
	CurrentPageURLRelative string // Relative URL, without "/documentation/<version>"
	CurrentPageURL         string // Full page URL
	MenuDocumentationLink  string
}

type versionMenuItems struct {
	Group      string
	Channel    string
	Version    string
	VersionURL string // Base URL for corresponding version without a leading /, e.g. 'v1.2.3-plus-fix6'.
	IsCurrent  bool
}

var ReleasesStatus ReleasesStatusType

var channelsListReverseStability = []string{"rock-solid", "stable", "ea", "beta", "alpha"}

func (m *versionMenuType) getChannelMenuData(r *http.Request, releases *ReleasesStatusType) (err error) {
	err = nil

	m.CurrentPageURLRelative = getDocPageURLRelative(r)
	m.CurrentPageURL = getCurrentPageURL(r)
	m.CurrentVersionURL = getVersionURL(r)

	if isGroupChannelURL, _ := regexp.MatchString("v[0-9]+.[0-9]+-(alpha|beta|ea|stable|rock-solid)", m.CurrentVersionURL); isGroupChannelURL {
		items := strings.Split(m.CurrentVersionURL, "-")
		if len(items) == 2 {
			m.CurrentGroup = strings.TrimPrefix(items[0], "v")
			m.CurrentChannel = items[1]
			_, m.CurrentVersion = getVersionFromChannelAndGroup(releases, m.CurrentChannel, m.CurrentGroup)
			m.CurrentVersionURL = VersionToURL(m.CurrentVersion)
		}
	} else {
		m.CurrentVersion = URLToVersion(m.CurrentVersionURL)
	}

	m.CurrentVersion = URLToVersion(m.CurrentVersionURL)

	if m.CurrentVersion == "" {
		m.CurrentVersion = fmt.Sprintf("v%s", getRootRelease())
		m.CurrentVersionURL = VersionToURL(m.CurrentVersion)
	}

	// Try to find current channel from URL
	if m.CurrentChannel == "" || m.CurrentGroup == "" {
		m.CurrentChannel, m.CurrentGroup = getChannelAndGroupFromVersion(releases, m.CurrentVersion)
	}

	// Add the first menu item
	m.VersionItems = append(m.VersionItems, versionMenuItems{
		Group:      m.CurrentGroup,
		Channel:    m.CurrentChannel,
		Version:    m.CurrentVersion,
		VersionURL: m.CurrentVersionURL,
		IsCurrent:  true,
	})

	// Add other items
	for _, group := range getGroups() {
		// TODO error handling
		_ = m.getChannelsFromGroup(&ReleasesStatus, group)
	}

	return
}

func (m *versionMenuType) getVersionMenuData(r *http.Request, releases *ReleasesStatusType) (err error) {
	err = nil

	m.CurrentPageURLRelative = getDocPageURLRelative(r)
	m.CurrentPageURL = getCurrentPageURL(r)
	m.CurrentVersionURL = getVersionURL(r)
	m.CurrentVersion = URLToVersion(m.CurrentVersionURL)

	if m.CurrentVersion == "" {
		m.CurrentVersion = fmt.Sprintf("v%s", getRootRelease())
		m.CurrentVersionURL = VersionToURL(m.CurrentVersion)
	}

	re := regexp.MustCompile(`^v([0-9]+\.[0-9]+)(\..+)?$`)
	res := re.FindStringSubmatch(m.CurrentVersion)
	if res == nil {
		m.MenuDocumentationLink = fmt.Sprintf("/documentation/%s/", VersionToURL(m.CurrentVersion))
	} else {
		if res[2] != "" {
			// Version is not a group (MAJ.MIN), but the patch version
			m.MenuDocumentationLink = fmt.Sprintf("/documentation/%s/", VersionToURL(res[1]))
			m.AbsoluteVersion = fmt.Sprintf("%s", m.CurrentVersion)
		} else {
			m.MenuDocumentationLink = fmt.Sprintf("/documentation/%s/", VersionToURL(m.CurrentVersion))
			err, m.AbsoluteVersion = getVersionFromGroup(&ReleasesStatus, res[1])
			if err != nil {
				log.Debugln(fmt.Sprintf("getVersionMenuData: error determine absolute version for %s (got %s)", m.CurrentVersion, m.AbsoluteVersion))
			}
		}
	}

	//m.MenuDocumentationLink = fmt.Sprintf("/documentation/v%s/", VersionToURL(getRootRelease()))
	//if m.CurrentChannel == "" && m.CurrentGroup == "" {
	//	m.MenuDocumentationLink = fmt.Sprintf("/documentation/%v/", VersionToURL(m.CurrentVersion))
	//} else if  m.CurrentChannel == "" && m.CurrentGroup != "" {
	//	m.MenuDocumentationLink = fmt.Sprintf("/documentation/v%v/", m.CurrentGroup)
	//} else {
	//	m.MenuDocumentationLink = fmt.Sprintf("/documentation/v%v-%v/", m.CurrentGroup, m.CurrentChannel)
	//}

	// Add the first menu item
	m.VersionItems = append(m.VersionItems, versionMenuItems{
		Group:      m.CurrentGroup,
		Channel:    m.CurrentChannel,
		Version:    m.CurrentVersion,
		VersionURL: m.CurrentVersionURL,
		IsCurrent:  true,
	})

	//for _, releaseItem_ := range releases.Releases {
	//	if releaseItem_.Group == m.CurrentGroup {
	//		for _, channelItem_ := range releaseItem_.Channels {
	//			if channelItem_.Name == m.CurrentChannel {
	//				m.VersionItems = append(m.VersionItems, versionMenuItems{
	//					Group:      m.CurrentGroup,
	//					Channel:    m.CurrentChannel,
	//					Version:    channelItem_.Version,
	//					VersionURL: VersionToURL(channelItem_.Version),
	//					IsCurrent:  true,
	//				})
	//			}
	//		}
	//	}
	//}

	// Add other items
	for _, group := range getGroups() {
		// TODO error handling
		_ = m.getChannelsFromGroup(&ReleasesStatus, group)
	}

	return
}

func (m *versionMenuType) getGroupMenuData(r *http.Request, releases *ReleasesStatusType) (err error) {
	err = nil

	m.CurrentPageURLRelative = getDocPageURLRelative(r)
	m.CurrentPageURL = getCurrentPageURL(r)
	m.CurrentVersionURL = getVersionURL(r)
	m.CurrentVersion = URLToVersion(m.CurrentVersionURL)

	if m.CurrentVersion == "" {
		m.CurrentVersion = fmt.Sprintf("v%s", getRootRelease())
		m.CurrentVersionURL = VersionToURL(m.CurrentVersion)
	}

	re := regexp.MustCompile(`^v([0-9]+\.[0-9]+)$`)
	res := re.FindStringSubmatch(m.CurrentVersion)
	if res != nil {
		m.VersionItems = append(m.VersionItems, versionMenuItems{
			Group:      res[1],
			Channel:    "",
			Version:    m.CurrentVersion,
			VersionURL: m.CurrentVersionURL,
			IsCurrent:  true,
		})
	} else {
		// Version is not a group (MAJ.MIN), but the patch version
		m.VersionItems = append(m.VersionItems, versionMenuItems{
			Group:      "",
			Channel:    "",
			Version:    m.CurrentVersion,
			VersionURL: m.CurrentVersionURL,
			IsCurrent:  true,
		})
	}

	// Add other items
	for _, group := range getGroups() {
		// TODO error handling
		if group == "1.0" {
			continue
		}
		m.VersionItems = append(m.VersionItems, versionMenuItems{
			Group:      group,
			Channel:    "",
			Version:    "",
			VersionURL: "",
			IsCurrent:  false,
		})
	}

	return
}

// Get channels and corresponding versions for the specified
// group according to the reverse order of stability
func (m *versionMenuType) getChannelsFromGroup(releases *ReleasesStatusType, group string) (err error) {
	if group == "1.0" {
		return
	}
	for _, item := range releases.Releases {
		if item.Group == group {
			for _, channel := range channelsListReverseStability {
				for _, channelItem := range item.Channels {
					if channelItem.Name == channel {
						m.VersionItems = append(m.VersionItems, versionMenuItems{
							Group:      group,
							Channel:    channelItem.Name,
							Version:    channelItem.Version,
							VersionURL: VersionToURL(channelItem.Version),
							IsCurrent:  false,
						})
					}
				}
			}
		}
	}
	return
}

// Get channel and group for specified version
func getChannelAndGroupFromVersion(releases *ReleasesStatusType, version string) (channel, group string) {

	re := regexp.MustCompile(`^v([0-9]+\.[0-9]+)$`)
	res := re.FindStringSubmatch(version)
	if res != nil {
		return "", res[1]
	}

	for _, group := range getGroups() {
		for _, channel := range channelsListReverseStability {
			for _, releaseItem := range releases.Releases {
				if releaseItem.Group == group {
					for _, channelItem := range releaseItem.Channels {
						if channelItem.Name == channel {
							if channelItem.Version == version {
								return channel, group
							}
						}
					}
				}
			}
		}
	}
	return
}

// Get version for specified group and channel
func getVersionFromChannelAndGroup(releases *ReleasesStatusType, channel, group string) (err error, version string) {
	for _, releaseItem := range releases.Releases {
		if releaseItem.Group == group {
			for _, channelItem := range releaseItem.Channels {
				if channelItem.Name == channel {
					return nil, normalizeVersion(channelItem.Version)
				}
			}
		}
	}
	return errors.New(fmt.Sprintf("no matching version for group %s, channel %s", group, channel)), ""
}

// Gev version from specified group
// E.g. get v1.2.3+fix6 from v1.2
func getVersionFromGroup(releases *ReleasesStatusType, group string) (err error, version string) {
	if len(releases.Releases) > 0 {
		for _, ReleaseGroup := range releases.Releases {
			if ReleaseGroup.Group == group {
				releaseVersions := make(map[string]string)
				for _, channel := range ReleaseGroup.Channels {
					releaseVersions[channel.Name] = channel.Version
				}

				if _, ok := releaseVersions["stable"]; ok {
					return nil, normalizeVersion(releaseVersions["stable"])
				} else if _, ok := releaseVersions["ea"]; ok {
					return nil, normalizeVersion(releaseVersions["ea"])
				} else if _, ok := releaseVersions["beta"]; ok {
					return nil, normalizeVersion(releaseVersions["beta"])
				} else if _, ok := releaseVersions["alpha"]; ok {
					return nil, (releaseVersions["alpha"])
				}
			}
		}
	}

	return errors.New(fmt.Sprintf("Can't get version for %s", group)), ""

}

// Add prefix 'v' to a version if it doesn't have yet
func normalizeVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	} else {
		return fmt.Sprintf("v%s", version)
	}
}

func getRootReleaseVersion() string {
	activeRelease := getRootRelease()

	_ = updateReleasesStatus()

	if len(ReleasesStatus.Releases) > 0 {
		for _, ReleaseGroup := range ReleasesStatus.Releases {
			if ReleaseGroup.Group == activeRelease {
				releaseVersions := make(map[string]string)
				for _, channel := range ReleaseGroup.Channels {
					releaseVersions[channel.Name] = channel.Version
				}

				if _, ok := releaseVersions["stable"]; ok {
					return releaseVersions["stable"]
				} else if _, ok := releaseVersions["ea"]; ok {
					return releaseVersions["ea"]
				} else if _, ok := releaseVersions["beta"]; ok {
					return releaseVersions["beta"]
				} else if _, ok := releaseVersions["alpha"]; ok {
					return releaseVersions["alpha"]
				}
			}
		}
	}
	return "unknown"
}

func getRootRelease() (result string) {

	if len(os.Getenv("ACTIVE_RELEASE")) > 0 {
		result = os.Getenv("ACTIVE_RELEASE")
	} else {
		result = "1.2"
	}

	return
}

func getPage(filename string) ([]byte, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Get the full page URL menu requested for
// E.g /documentation/v1.2.3/reference/build_process.html
func getCurrentPageURL(r *http.Request) (result string) {

	originalURI, err := url.Parse(r.Header.Get("x-original-uri"))
	if err != nil {
		return
	}

	if originalURI.Path == "/404.html" {
		return
	} else {
		return originalURI.Path
	}

}

// Get page URL menu requested for without a leading version suffix
// E.g /reference/build_process.html for /documentation/v1.2.3/reference/build_process.html
// if useURI == true - use requestURI instead of x-original-uri header value
func getDocPageURLRelative(r *http.Request, useURI ...bool) (result string) {
	var (
		URLtoParse  string
		originalURI *url.URL
		err         error
	)

	if len(useURI) > 0 && useURI[0] {
		originalURI, err = url.Parse(r.RequestURI)
	} else {
		originalURI, err = url.Parse(r.Header.Get("x-original-uri"))
	}

	if err != nil {
		return
	}

	if originalURI.Path == "/404.html" {
		return
	} else {
		URLtoParse = originalURI.Path
	}

	re := regexp.MustCompile(`^/documentation/[^/]+(/.+)$`)
	res := re.FindStringSubmatch(URLtoParse)
	if res != nil {
		result = res[1]
	}
	return
}

// Get version URL page belongs to if request came from concrete documentation version, otherwise empty.
// E.g for the /documentation/v1.2.3-plus-fix5/reference/build_process.html return "v1.2.3-plus-fix5".
func getVersionURL(r *http.Request) (result string) {
	URLtoParse := ""
	originalURI, err := url.Parse(r.Header.Get("x-original-uri"))

	if err != nil {
		return
	}

	if originalURI.Path == "/404.html" {
		values, err := url.ParseQuery(originalURI.RawQuery)
		if err != nil {
			return
		}
		URLtoParse = values.Get("uri")
	} else {
		URLtoParse = originalURI.Path
	}

	re := regexp.MustCompile(`^/documentation/([^/]+)/?.*$`)
	res := re.FindStringSubmatch(URLtoParse)
	if res != nil {
		result = res[1]
	}

	return strings.TrimPrefix(result, "/")
}

func inspectRequest(r *http.Request) string {
	var request []string

	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		_ = r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}

	return strings.Join(request, "\n")
}

func inspectRequestHTML(r *http.Request) string {
	var request []string

	request = append(request, "<p>")
	url := fmt.Sprintf("<b>%v</b> %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	request = append(request, fmt.Sprintf("<b>Host:</b> %v", r.Host))
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("<b>%v:</b> %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		_ = r.ParseForm()
		request = append(request, r.Form.Encode())
	}

	request = append(request, "</p>")
	return strings.Join(request, "<br />")
}

func VersionToURL(version string) string {
	result := strings.ReplaceAll(version, "+", "-plus-")
	result = strings.ReplaceAll(result, "_", "-u-")
	return normalizeVersion(result)
}

func URLToVersion(version string) (result string) {
	result = strings.ReplaceAll(version, "-plus-", "+")
	result = strings.ReplaceAll(result, "-u-", "_")
	return
}

func validateURL(URL string) (err error) {
	if strings.ToLower(os.Getenv("URL_VALIDATION")) == "false" {
		return nil
	}

	var resp *http.Response
	allowedStatusCodes := []int{200, 401}
	tries := 3
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for {
		resp, err = client.Get(URL)
		log.Tracef("Validating %v (tries-%v):\nStatus - %v\nHeader - %+v,", URL, tries, resp.Status, resp.Header)
		if err == nil && (resp.StatusCode == 301 || resp.StatusCode == 302) {
			if len(resp.Header.Get("Location")) > 0 {
				URL = resp.Header.Get("Location")
			} else {
				tries = 0
			}
			tries--
		} else {
			tries = 0
		}
		if tries < 1 {
			break
		}
	}

	if err == nil {
		place := sort.SearchInts(allowedStatusCodes, resp.StatusCode)
		if place >= len(allowedStatusCodes) {
			err = errors.New(fmt.Sprintf("%s is not valid", URL))
		}
	}
	return
}

// Get update channel groups in a descending order.
func getGroups() (groups []string) {
	for _, item := range ReleasesStatus.Releases {
		groups = append(groups, item.Group)
	}
	sort.Slice(groups, func(i, j int) bool {
		var i_, j_ float64
		var err error
		if i_, err = strconv.ParseFloat(groups[i], 32); err != nil {
			i_ = 0
		}
		if j_, err = strconv.ParseFloat(groups[j], 32); err != nil {
			j_ = 0
		}
		return i_ > j_
	})
	return
}

func getRootFilesPath(r *http.Request) (result string) {
	result = "./root/"
	if strings.HasPrefix(r.Host, "ru.") {
		result += "ru"
	} else {
		result += "main"
	}
	return
}

func updateReleasesStatus() error {
	err := updateReleasesStatusTRDL()
	return err
}

func updateReleasesStatusMultiwerf() error {
	data, err := ioutil.ReadFile("multiwerf/multiwerf.json")
	if err != nil {
		log.Errorf("Can't open multiwerf.json (%e)", err)
		return err
	}
	err = json.Unmarshal(data, &ReleasesStatus)
	if err != nil {
		log.Errorf("Can't unmarshall multiwerf.json (%e)", err)
		return err
	}
	return err
}

func updateReleasesStatusTRDL() error {
	data, err := ioutil.ReadFile("trdl/trdl_channels.yaml")
	if err != nil {
		log.Errorf("Can't open trdl_channels.yaml (%e)", err)
		return err
	}
	err = yaml.Unmarshal(data, &ReleasesStatus)
	if err != nil {
		log.Errorf("Can't unmarshall trdl_channels.yaml (%e)", err)
		return err
	}
	return err
}
