package suite_init

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/prashantv/gostub"

	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/test/pkg/utils"
)

type ProjectNameData struct {
	ProjectName string
}

func NewProjectNameData(stubsData *StubsData) *ProjectNameData {
	data := &ProjectNameData{}
	SetupProjectName(&data.ProjectName, stubsData.Stubs)
	return data
}

func SetupProjectName(projectName *string, stubs *gostub.Stubs) bool {
	return BeforeEach(func() {
		*projectName = GenerateUniqProjectName()
		stubs.SetEnv("WERF_PROJECT_NAME", *projectName)
	})
}

func GenerateUniqProjectName() string {
	var packageId string
	filename := filepath.Base(os.Args[0])
	if strings.HasPrefix(filename, "___") { // ide
		packageId = "none"
	} else {
		packageId = strings.Split(filename, ".")[0] // .test .test.exe
	}

	projectName := strings.Join([]string{
		"werf-test",
		packageId,
		strconv.Itoa(os.Getpid()),
		utils.GetRandomString(10),
	}, "-")

	return slug.LimitedSlug(projectName, 30)
}
