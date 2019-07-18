---
title: SSH
sidebar: documentation
permalink: documentation/reference/toolbox/ssh.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

There are cases, when werf might need ssh-keys of the user:

1. To clone remote git repos, which can be specified in the werf.yaml configuration.
2. Images build instructions need to access external data over ssh.

By default (without any options) werf will try to use system ssh-agent by checking `SSH_AUTH_SOCK` environment variable.

In the case, when there is no system ssh-agent running, werf try to behave like ssh-client: use default ssh-keys of user which run werf (`~/.ssh/id_rsa|id_dsa`). If werf detects one of these files, then temporary ssh-agent will be run by werf and these keys will be added into this agent.

To enable only specific ssh-keys use option `--ssh-key PRIVATE_KEY_FILE_PATH` (can be specified multiple times to use multiple keys). In this case werf will run temporary ssh-agent and add only specified keys into it.

## How werf works with ssh-agent

Auth sock `SSH_AUTH_SOCK` will be used when working with remote git repos and will be mounted into all build containers, so that build instructions can access this agent.

**NOTICE** There is a restriction, that only `root` user (default) inside build container can access ssh-agent by mounted `SSH_AUTH_SOCK`.

## Temporary ssh-agent

Werf can start such an agent in some command. This ssh-agent process will terminate as werf command exits. Temporary ssh-agent process do not conflicts with default system ssh-agent in the case there is one.
