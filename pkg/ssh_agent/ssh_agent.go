package ssh_agent

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

var (
	SSHAuthSock string
	tmpSockPath string
)

func Init(keys []string) error {
	for _, key := range keys {
		if keyExists, err := util.FileExists(key); !keyExists {
			return fmt.Errorf("specified ssh key %s does not exist", key)
		} else if err != nil {
			return fmt.Errorf("specified ssh key %s does not exist: %v", key, err)
		}
	}

	if len(keys) > 0 {
		agentSock, err := runSSHAgentWithKeys(keys)
		if err != nil {
			return err
		}
		SSHAuthSock = agentSock

		return nil
	}

	systemAgentSock := os.Getenv("SSH_AUTH_SOCK")
	systemAgentSockExists, _ := util.FileExists(systemAgentSock)
	if systemAgentSock != "" && systemAgentSockExists {
		SSHAuthSock = systemAgentSock
		logboek.LogF("Using system ssh-agent: %s\n", systemAgentSock)
		return nil
	}

	var defaultKeys []string
	for _, defaultFileName := range []string{"id_rsa", "id_dsa"} {
		path := filepath.Join(os.Getenv("HOME"), ".ssh", defaultFileName)

		if keyExists, _ := util.FileExists(path); keyExists {
			defaultKeys = append(defaultKeys, path)
		}
	}

	if len(defaultKeys) > 0 {
		var validKeys []string

		for _, key := range defaultKeys {
			keyData, err := ioutil.ReadFile(key)
			if err != nil {
				logboek.Warn.LogF("WARNING: cannot read default key %s: %s\n", key, err)
				continue
			}
			_, err = ssh.ParseRawPrivateKey(keyData)
			if err != nil {
				logboek.Warn.LogF("WARNING: default key %s validation error: %s\n", key, err)
				continue
			}

			validKeys = append(validKeys, key)
		}

		if len(validKeys) > 0 {
			agentSock, err := runSSHAgentWithKeys(validKeys)
			if err != nil {
				return err
			}
			SSHAuthSock = agentSock
		}
	}

	return nil
}

func Terminate() error {
	if tmpSockPath != "" {
		err := os.RemoveAll(tmpSockPath)
		if err != nil {
			return fmt.Errorf("unable to remove tmp ssh agent sock %s: %s", tmpSockPath, err)
		}
	}

	return nil
}

func runSSHAgentWithKeys(keys []string) (string, error) {
	agentSock, err := runSSHAgent()
	if err != nil {
		return "", fmt.Errorf("error running ssh agent: %s", err)
	}

	for _, key := range keys {
		err := addSSHKey(agentSock, key)
		if err != nil {
			return "", fmt.Errorf("error adding ssh key %s: %s", key, err)
		}
	}

	return agentSock, nil
}

func runSSHAgent() (string, error) {
	sockPath := filepath.Join(werf.GetTmpDir(), "werf-ssh-agent", uuid.NewV4().String())
	tmpSockPath = sockPath

	err := os.MkdirAll(filepath.Dir(sockPath), os.ModePerm)
	if err != nil {
		return "", err
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return "", fmt.Errorf("error listen unix sock %s: %s", sockPath, err)
	}

	logboek.LogF("Running ssh agent on unix sock: %s\n", sockPath)

	go func() {
		agnt := agent.NewKeyring()

		for {
			conn, err := ln.Accept()
			if err != nil {
				logboek.LogWarnF("WARNING: failed to accept ssh-agent connection: %s\n", err)
				continue
			}

			go func() {
				var err error

				err = agent.ServeAgent(agnt, conn)
				if err != nil && err != io.EOF {
					logboek.LogWarnF("WARNING: ssh-agent server error: %s\n", err)
					return
				}

				err = conn.Close()
				if err != nil {
					logboek.LogWarnF("WARNING: ssh-agent server connection close error: %s\n", err)
					return
				}
			}()
		}
	}()

	return sockPath, nil
}

func addSSHKey(authSock string, key string) error {
	conn, err := net.Dial("unix", authSock)
	if err != nil {
		return fmt.Errorf("error dialing with ssh agent %s: %s", authSock, err)
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)

	keyData, err := ioutil.ReadFile(key)
	if err != nil {
		return fmt.Errorf("error reading key file %s: %s", key, err)
	}

	privateKey, err := ssh.ParseRawPrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("error parsing private key %s: %s", key, err)
	}

	err = agentClient.Add(agent.AddedKey{PrivateKey: privateKey})
	if err != nil {
		return err
	}

	logboek.LogF("Added private key %s to ssh agent %s\n", key, authSock)

	return nil
}
