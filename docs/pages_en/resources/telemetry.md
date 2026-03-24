---
title: Telemetry
permalink: resources/telemetry.html
---

We collect anonymous usage data to improve werf's features and steer its development in the right direction. This data is not associated with users in any way and does not contain any personal information.

It helps us figure out how werf is used and focus on improving the features that are most needed.

## Example of the data werf transmits and its breakdown

Below are examples of the data werf transmits:

```json
{
  "ts": 1658231825280,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "userID": "",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab",
    "extra": {}
  },
  "eventType": "CommandStarted",
  "eventData": {
    "commandOptions": [
      {
        "name": "repo",
        "asCli": false,
        "asEnv": false,
        "count": 0
      }
    ]
  },
  "schemaVersion": 2
}
```

```json
{
  "ts": 1658231836102,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "userID": "",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab",
    "extra": {}
  },
  "eventType": "CommandExited",
  "eventData": {
    "exitCode": 0,
    "durationMs": 10827
  },
  "schemaVersion": 2
}
```

Example of build started event:

```json
{
  "ts": 1709251100000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "BuildStarted",
  "eventData": {
    "imagesCount": 3,
    "containerBackend": "docker",
    "inContainer": false
  },
  "schemaVersion": 2
}
```

Example of build finished event:

```json
{
  "ts": 1709251150000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "BuildFinished",
  "eventData": {
    "durationMs": 50000,
    "success": true,
    "imagesCount": 3
  },
  "schemaVersion": 2
}
```

Example of image build finished event:

```json
{
  "ts": 1709251140000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "ImageBuildFinished",
  "eventData": {
    "image": "backend",
    "durationMs": 15500,
    "rebuilt": true,
    "configType": "stapel"
  },
  "schemaVersion": 2
}
```

Example of stage build finished event:

```json
{
  "ts": 1709251135000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "StageBuildFinished",
  "eventData": {
    "image": "backend",
    "stage": "install",
    "durationMs": 8500,
    "fromCache": false,
    "baseImageSource": "repo",
    "baseImagePulled": true
  },
  "schemaVersion": 2
}
```

In the examples above, the following usage data is sent:

* `ts` — timestamp;
* `executionID` — run UUID;
* `userID` — anonymous user identifier; may be empty if not set;
* `projectID` — checksum (SHA-256) of the Git origin URL of the project;
* `command` — the name of the werf command being executed;
* `attributes` — environment attributes:
  * `os` — operating system;
  * `arch` — CPU architecture;
  * `groupChannel` — selected `trdl` group channel if set via `TRDL_USE_WERF_GROUP_CHANNEL`;
  * `version` — werf version;
  * `ci` — whether a CI/CD system is used;
  * `ciName` — detected CI/CD system name (`gitlab`, `github-actions`, etc.);
  * `extra` — additional anonymous attributes from environment variables with the `WERF_TELEMETRY_EXTRA_ATTRIBUTE_*` prefix.
* `eventType` — type of event:
  * `CommandStarted` — command execution started;
  * `CommandExited` — command execution completed;
  * `UnshallowFailed` — error during `unshallow` in GitLab CI;
  * `BuildStarted` — image build started;
  * `BuildFinished` — image build completed;
  * `ImageBuildFinished` — individual image build completed;
  * `StageBuildFinished` — individual stage build completed.
* `eventData` — event data specific to each event type:
  * for `CommandStarted`: list of command options with flags (`asCli`, `asEnv`, `count`);
  * for `CommandExited`: exit code (`exitCode`) and command execution duration in milliseconds (`durationMs`);
  * for `UnshallowFailed`: error message (`errorMessage`), GitLab Runner version (`gitlabRunnerVersion`), and GitLab Server version (`gitlabServerVersion`);
  * for `BuildStarted`: number of images to build (`imagesCount`), selected container backend (`containerBackend`: `docker`/`buildah`), and whether werf runs inside a container (`inContainer`);
  * for `BuildFinished`: build duration (`durationMs`), success status (`success`), and number of images (`imagesCount`);
  * for `ImageBuildFinished`: image name (`image`), build time (`durationMs`), rebuild flag (`rebuilt`), and image config type (`configType`: `stapel`/`dockerfile`/`staged`/`unknown`);
  * for `StageBuildFinished`: image name (`image`), stage name (`stage`), build time (`durationMs`), cache usage flag (`fromCache`), base image source (`baseImageSource`: `repo`/`secondary`/empty value), and whether the base image was pulled (`baseImagePulled`).

We collect only anonymized information, as shown in the examples above. You can verify this yourself in the werf source code responsible for telemetry: the [event.go](https://github.com/werf/werf/blob/main/pkg/telemetry/event.go) and [telemetrywerfio.go](https://github.com/werf/werf/blob/main/pkg/telemetry/telemetrywerfio.go) files in the [telemetry](https://github.com/werf/werf/tree/main/pkg/telemetry) package.

## Configuring the telemetry log file

We trust our users and want them to trust us, so we tried to make the process as transparent as possible. Use the `WERF_TELEMETRY_LOG_FILE` environment variable to specify the path to the log file where all telemetry data will be stored.

## Disabling telemetry

You can disable telemetry by setting the `WERF_TELEMETRY` environment variable to `0`:

```shell
export WERF_TELEMETRY=0
```
