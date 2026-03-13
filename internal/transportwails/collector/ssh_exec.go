package collector

import (
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"golang.org/x/crypto/ssh"
)

func runSSHCommand(config *connection.SSHTunnelConfig, cmd string) (string, error) {
	sshConfig, err := sshClientConfig(config)
	if err != nil {
		return "", err
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func sshClientConfig(config *connection.SSHTunnelConfig) (*ssh.ClientConfig, error) {
	clientConfig := &ssh.ClientConfig{
		User:            config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	if config.Password != "" {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(config.Password))
	}
	if len(clientConfig.Auth) == 0 {
		return nil, fmt.Errorf("ssh password is required")
	}
	return clientConfig, nil
}
