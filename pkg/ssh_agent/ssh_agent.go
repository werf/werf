package ssh_agent

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/logger"
	"github.com/flant/dapp/pkg/util"
)

var (
	SSHAuthSock string
)

func Init(keys []string) error {
	for _, key := range keys {
		if !util.IsFileExists(key) {
			return fmt.Errorf("specified ssh key %s does not exist", key)
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
	if systemAgentSock != "" && util.IsFileExists(systemAgentSock) {
		SSHAuthSock = systemAgentSock
		fmt.Printf("Using system ssh-agent %s\n", systemAgentSock)
		return nil
	}

	var defaultKeys []string
	for _, defaultFileName := range []string{"id_rsa", "id_dsa"} {
		path := filepath.Join(os.Getenv("HOME"), ".ssh", defaultFileName)
		if util.IsFileExists(path) {
			defaultKeys = append(defaultKeys, path)
		}
	}

	if len(defaultKeys) > 0 {
		var validKeys []string

		for _, key := range defaultKeys {
			keyData, err := ioutil.ReadFile(key)
			if err != nil {
				continue
			}
			_, err = ssh.ParseRawPrivateKeyWithPassphrase(keyData, []byte{})
			if err != nil {
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
	sockPath := filepath.Join(dapp.GetTmpDir(), "dapp-ssh-agent", uuid.NewV4().String())

	err := os.MkdirAll(filepath.Dir(sockPath), os.ModePerm)
	if err != nil {
		return "", err
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return "", fmt.Errorf("error listen unix sock %s: %s", sockPath, err)
	}

	fmt.Printf("Running ssh agent on unix sock %s\n", sockPath)

	go func() {
		agnt := agent.NewKeyring()

		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.LogWarningF("WARNING: failed to accept ssh-agent connection: %s\n", err)
				continue
			}

			go func() {
				var err error

				err = agent.ServeAgent(agnt, conn)
				if err != nil && err != io.EOF {
					logger.LogWarningF("WARNING: ssh-agent server error: %s\n", err)
					return
				}

				err = conn.Close()
				if err != nil {
					logger.LogWarningF("WARNING: ssh-agent server connection close error: %s\n", err)
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

	privateKey, err := ssh.ParseRawPrivateKeyWithPassphrase(keyData, []byte{})
	if err != nil {
		return fmt.Errorf("error parsing private key %s: %s", key, err)
	}

	err = agentClient.Add(agent.AddedKey{PrivateKey: privateKey})
	if err != nil {
		return err
	}

	fmt.Printf("Added private key %s to ssh agent %s\n", key, authSock)

	return nil
}
