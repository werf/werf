---
title: Integration with SSH-agent
permalink: internals/integration_with_ssh_agent.html
---

werf might require user ssh-keys in the following cases:

1. To clone remote git repositories that are specified in the werf.yaml configuration.
2. To provide access to external data over ssh for instructions for building images.

By default, (without any options) werf will try to use the system ssh-agent by checking `SSH_AUTH_SOCK` environment variable.

If there is no ssh-agent running in the system, werf will try to behave like a ssh-client by using the default ssh-keys of the user that runs werf (i.e., `~/.ssh/id_rsa|id_dsa`). If werf detects one of these files, it starts a temporary ssh-agent and adds these keys to it.

You can enable only specific ssh-keys by setting an option `--ssh-key PRIVATE_KEY_FILE_PATH` (it can be set multiple times to use numerous keys). In this case, werf will run a temporary ssh-agent and add the specified keys to it .

## How werf works with a ssh-agent

When working with remote git repos, werf uses the `SSH_AUTH_SOCK` environment variable that contains the path of the UNIX file socket. This socket is mounted into all building containers so that build instructions can access an agent.

**NOTE** Only a `root` (default) user can access ssh-agent via the mounted `SSH_AUTH_SOCK` inside the building container.

## A temporary ssh-agent

werf might run a temporary ssh-agent for some commands to work correctly. Such ssh-agent terminates when a corresponding werf command finishes its work. A temporary ssh-agent does not conflict with the default ssh-agent if one is present in the system.
