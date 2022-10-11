package compose

import (
	"fmt"

	"github.com/werf/werf/cmd/werf/docs/structs"
)

func GetComposeDocs(short string) structs.DocsStruct {
	var docs structs.DocsStruct

	docs.Long = short
	docs.Long += `
Image environment name format: $WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME ($WERF_DOCKER_IMAGE_NAME for nameless image).
<FORMATTED_WERF_IMAGE_NAME> is werf image name from werf.yaml modified according to the following rules:
- all characters are uppercase (app -> APP);
- charset /- is replaced with _ (DEV/APP-FRONTEND -> DEV_APP_FRONTEND).
If one or more IMAGE_NAME parameters specified, werf will build and forward only these images.
Given the following werf configuration:
# werf.yaml
project: x
configVersion: 1
---
image: frontend
dockerfile: frontend.Dockerfile
---
image: geodata-backend
dockerfile: backend.Dockerfile
Use described images as follows in your docker compose configuration:
# docker-compose.yaml
services:
  frontend:
    image: $WERF_FRONTEND_DOCKER_IMAGE_NAME
  backend:
    image: $WERF_GEODATA_BACKEND_DOCKER_IMAGE_NAME
`

	docs.LongMD = short + "\n\n" +
		"Image environment name format: `$WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME`:\n" +
		"* `$WERF_DOCKER_IMAGE_NAME` for nameless image;\n" +
		"* `<FORMATTED_WERF_IMAGE_NAME>` is werf image name from `werf.yaml` modified according to the following\n" +
		"rules:\n" +
		"  * all characters are uppercase (`app` -&gt; `APP`);\n" +
		"  * charset `/` - is replaced with `_` (`DEV/APP-FRONTEND` -&gt; `DEV_APP_FRONTEND`).\n\n" +
		"If one or more `IMAGE_NAME` parameters specified, werf will build and forward only these images.\n\n" +
		"Given the following werf configuration:\n\n" +
		"## werf.yaml\n" +
		"```shell\n" +
		"project: x\n" +
		"configVersion: 1\n" +
		"---\n" +
		"image: frontend\n" +
		"dockerfile: frontend.Dockerfile\n" +
		"---\n" +
		"image: geodata-backend\n" +
		"dockerfile: backend.Dockerfile\n" +
		"```\n\n" +
		"Use described images as follows in your docker compose configuration:\n\n" +
		"## docker-compose.yaml\n" +
		"```shell\n" +
		"services:\n" +
		"  frontend:\n" +
		"    image: $WERF_FRONTEND_DOCKER_IMAGE_NAME\n" +
		"  backend:\n" +
		"    image: $WERF_GEODATA_BACKEND_DOCKER_IMAGE_NAME\n" +
		"```\n"

	return docs
}

func GetComposeShort(composeCmdName string) structs.DocsShortStruct {
	var shorts structs.DocsShortStruct

	shorts.Short = fmt.Sprintf("Run docker-compose %s command with forwarded image names.", composeCmdName)
	shorts.ShortMD = fmt.Sprintf("Run `docker-compose %s` command with forwarded image names.", composeCmdName)

	return shorts
}
