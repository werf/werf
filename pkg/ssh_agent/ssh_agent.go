package ssh_agent

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/flant/dapp/pkg/dapp"
	uuid "github.com/satori/go.uuid"

	"github.com/flant/dapp/pkg/util"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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
		agentSock, err := runSSHAgent()
		if err != nil {
			return fmt.Errorf("error running ssh agent: %s", err)
		}
		SSHAuthSock = agentSock

		fmt.Printf("Running ssh agent on unix sock %s\n", SSHAuthSock)

		for _, key := range keys {
			err := addSSHKey(agentSock, key)
			if err != nil {
				return fmt.Errorf("error adding ssh key %s: %s", key, err)
			}
		}

		return nil
	}

	systemAgentSock := os.Getenv("SSH_AUTH_SOCK")
	if systemAgentSock != "" && util.IsFileExists(systemAgentSock) {
		SSHAuthSock = systemAgentSock
		fmt.Printf("Using system ssh-agent %s\n", systemAgentSock)
		return nil
	}

	defaultKeys := []string{}
	for _, defaultFileName := range []string{"id_rsa", "id_dsa"} {
		path := filepath.Join(os.Getenv("HOME"), ".ssh", defaultFileName)
		if util.IsFileExists(path) {
			defaultKeys = append(defaultKeys, path)
		}
	}

	if len(defaultKeys) > 0 {
		agentSock, err := runSSHAgent()
		if err != nil {
			return fmt.Errorf("error running ssh agent: %s", err)
		}
		SSHAuthSock = agentSock

		fmt.Printf("Running ssh agent on unix sock %s\n", SSHAuthSock)

		for _, key := range defaultKeys {
			// TODO: askpass
			err := addSSHKey(agentSock, key)
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARNING failed to add default ssh key %s to ssh-agent: %s\n", key, err)
			}
		}
	}

	return nil
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

	go func() {
		agnt := agent.NewKeyring()

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARNING: error accepting connection: %s\n", err)
				continue
			}

			go func() {
				var err error

				err = agent.ServeAgent(agnt, conn)
				if err != nil && err != io.EOF {
					fmt.Fprintf(os.Stderr, "WARNING: ssh-agent server error: %s\n", err)
					return
				}

				err = conn.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARNING: ssh-agent server connection close error: %s\n", err)
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
