package main

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ChannelType struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ReleaseType struct {
	Group    string        `json:"group"`
	Channels []ChannelType `json:"channels"`
}

type ReleasesStatusType struct {
	Releases []ReleaseType `json:"multiwerf"`
}

type ApiStatusResponseType struct {
	Status         string        `json:"status"`
	Msg            string        `json:"msg"`
	RootVersion    string        `json:"rootVersion"`
	RootVersionURL string        `json:"rootVersionURL"`
	Multiwerf      []ReleaseType `json:"multiwerf"`
}

type versionMenuType struct {
	VersionItems          []versionMenuItems
	HTMLContent           string
	CurrentGroup          string
	CurrentChannel        string
	CurrentVersion        string
	CurrentVersionURL     string
	CurrentPageURL        string // Page URL menu requesting for with a leading /documentation/, e.g /documentation/reference/build_process.html. Or "/documentation/" for unknown cases.
	MenuDocumentationLink string
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

func (m *versionMenuType) getVersionMenuData(r *http.Request, releases *ReleasesStatusType) (err error) {
	err = nil

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

	if m.CurrentVersion == "" {
		m.CurrentVersion = getRootReleaseVersion()
		m.CurrentVersionURL = VersionToURL(m.CurrentVersion)
	}

	// Try to find current channel from URL
	if m.CurrentChannel == "" || m.CurrentGroup == "" {
		m.CurrentChannel, m.CurrentGroup = getChannelAndGroupFromVersion(releases, m.CurrentVersion)
	}

	if m.CurrentChannel == "" || m.CurrentGroup == "" {
		m.MenuDocumentationLink = fmt.Sprintf("/%v/documentation/", m.CurrentVersion)
	} else {
		m.MenuDocumentationLink = fmt.Sprintf("/v%v-%v/documentation/", m.CurrentGroup, m.CurrentChannel)
	}

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
		_ = m.getChannelsForGroup(group, &ReleasesStatus)
	}

	return
}

// Get channels and corresponding versions for the specified
// group according to the reverse order of stability
func (m *versionMenuType) getChannelsForGroup(group string, releases *ReleasesStatusType) (err error) {
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
					return nil, channelItem.Version
				}
			}
		}
	}
	return errors.New(fmt.Sprintf("no matching version for group %s, channel %s", group, channel)), ""
}

func getRootReleaseVersion() string {
	var activeRelease string

	if len(os.Getenv("ACTIVE_RELEASE")) > 0 {
		activeRelease = os.Getenv("ACTIVE_RELEASE")
	} else {
		activeRelease = "1.2"
	}

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

func getPage(filename string) ([]byte, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Get page URL menu requested for with a leading /documentation/
// E.g /documentation/reference/build_process.html. Or "/documentation/" for unknown cases.
func getCurrentPageURL(r *http.Request) (result string) {
	URLtoParse := ""
	result = "/documentation/"
	originalURI, err := url.Parse(r.Header.Get("x-original-uri"))

	if err != nil {
		return
	}

	if originalURI.Path == "/404.html" {
		return
	} else {
		URLtoParse = originalURI.Path
	}

	if strings.Contains(URLtoParse, "/documentation/") {
		items := strings.Split(URLtoParse, "/documentation/")
		if len(items) > 1 {
			result += strings.Join(items[1:], "/documentation/")
		}
	}
	return
}

// Get version URL page belongs to if request came from concrete documentation version, otherwise empty.
// E.g for /v1.2.3-plus-fix5/documentation/reference/build_process.html return "v1.2.3-plus-fix5".
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
	if strings.Contains(URLtoParse, "/documentation/") {
		result = strings.Split(URLtoParse, "/documentation/")[0]
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
	return result
}

func URLToVersion(version string) (result string) {
	result = strings.ReplaceAll(version, "-plus-", "+")
	result = strings.ReplaceAll(result, "-u-", "_")
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
