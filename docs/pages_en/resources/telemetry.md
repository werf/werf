---
title: Telemetry
permalink: resources/telemetry.html
---

We collect anonymous usage data to improve werf's features and steer its development in the right direction.  This data is not associated with users in any way and does not contain any personal information.

It helps us figure out how werf is used and focus on improving the features that are most needed.

## Example of the data werf transmits and its breakdown

Below are examples of the data werf transmits:

```json
{
  "ts": 1658231825280,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "CI": true,
    "ciName": "gitlab"
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
  "schemaVersion": 1
}
```

```json
{
  "ts": 1658231836102,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab"
  },
  "eventType": "CommandExited",
  "eventData": {
    "exitCode": 0,
    "durationMs": 10827
  },
  "schemaVersion": 1
}
```

In the examples above, the following usage data is sent:

* `ts` — timestamp;
* `executionID` — UUID;
* `projectID` — checksum (SHA-256) of the git origin URL of the project;
* `command` — the name of werf command to be executed;
* `attributes` — environment attributes:
  * `os`;
  * `arch`;
  * `trdl group-channel`;
  * `werf version`;
  * `ci` — whether CI-system is used;
  * `ciName` — detected name of CI/CD system (gitlab, github-actions, etc.).
* `eventType` — type of event:
  * `CommandStarted`;
  * `CommandExited`;
* `eventData` — event data that includes:
  * exit code;
  * command running time;
  * the names of the options used.

We collect only anonymized information, as shown in the examples above. You can see this for yourself by analyzing the werf source code for telemetry, namely the [event.go](https://github.com/werf/werf/blob/main/pkg/telemetry/event.go) and [telemetrywerfio.go](https://github.com/werf/werf/blob/main/pkg/telemetry/telemetrywerfio.go) files of the [telemetry](https://github.com/werf/werf/tree/main/pkg/telemetry) package.

## Configuring the telemetry log file

We trust our users and want them to trust us, so we tried to make the process as transparent as possible. Use the `WERF_TELEMETRY_LOG_FILE` environment variable to specify the path to the log file where all telemetry data will be stored.

## Disabling telemetry

You can disable telemetry by setting the `WERF_TELEMETRY` environment variable to `0`:

```shell
export WERF_TELEMETRY=0
``` 
