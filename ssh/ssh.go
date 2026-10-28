package ssh

import (
	"os"
	"fmt"
	"time"
	"net"
	"strconv"
	"strings"
	"path/filepath"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	default_port	= 22
	default_user	= "root"
	default_timeout	= 5
)

type (
	Config struct {
		Host		string
		Port		int
		User		string
		Private_key	string
		Passphrase	string
		Known_hosts	string
		Timeout		int
	}
	
	Client struct {
		client *ssh.Client
	}
)

func Connect(config Config) (*Client, error){
	if config.Host == "" {
		return nil, fmt.Errorf("SSH host must be defined")
	}
	if config.Private_key == "" {
		return nil, fmt.Errorf("SSH private key must be defined")
	}
	
	if config.Port == 0 {
		config.Port = default_port
	}
	if config.User == "" {
		config.User = default_user
	}
	if config.Known_hosts == "" {
		config.Known_hosts = filepath.Join(filepath.Dir(config.Private_key), "known_hosts")
	}
	if config.Timeout == 0 {
		config.Timeout = default_timeout
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("SSH timeout must not be negative")
	}
	
	signer, err := load_private_key(config.Private_key, config.Passphrase)
	if err != nil {
		return nil, err
	}
	
	host_key_callback, err := knownhosts.New(config.Known_hosts)
	if err != nil {
		return nil, fmt.Errorf("Unable to load SSH known hosts: %w", err)
	}
	
	address		:= net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	timeout		:= time.Duration(config.Timeout) * time.Second
	deadline	:= time.Now().Add(timeout)
	
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to %s: %w", address, err)
	}
	
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, fmt.Errorf("Unable to set SSH deadline: %w", err)
	}
	
	connection, channels, requests, err := ssh.NewClientConn(conn, address, &ssh.ClientConfig{
		User:				config.User,
		Auth:				[]ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback:	host_key_callback,
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Unable to establish SSH connection to %s: %w", address, err)
	}
	
	if err := conn.SetDeadline(time.Time{}); err != nil {
		connection.Close()
		return nil, fmt.Errorf("Unable to clear SSH deadline: %w", err)
	}
	
	client := ssh.NewClient(connection, channels, requests)
	
	return &Client{client: client}, nil
}

func (c *Client) Copy_file(local, remote string) error {
	file, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("Unable to open local file %s: %w", local, err)
	}
	defer file.Close()
	
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("Unable to create SSH session: %w", err)
	}
	defer session.Close()
	
	session.Stdin = file
	
	command := fmt.Sprintf(
		"umask 077 && cat > %s && chmod 0600 %s",
		Shell_quote(remote),
		Shell_quote(remote),
	)
	
	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf(
			"Unable to copy %s to %s: %w\n%s",
			local,
			remote,
			err,
			output,
		)
	}
	
	return nil
}

func (c *Client) Run(command string) (string, error){
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("Unable to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf(
			"Remote command failed: %w\n%s",
			err,
			output,
		)
	}

	return string(output), nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func Shell_quote(value string) string {
	return "'"+strings.ReplaceAll(value, "'", "'\"'\"'")+"'"
}

func load_private_key(path, passphrase string) (ssh.Signer, error){
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read SSH private key %s: %w", path, err)
	}
	
	var signer ssh.Signer
	if passphrase == "" {
		signer, err = ssh.ParsePrivateKey(key)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to parse SSH private key %s: %w", path, err)
	}
	return signer, nil
}