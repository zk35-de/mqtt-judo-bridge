package dcm

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

// Client connects to the Judo DevCommManagerDaemon on TCP port 8833.
// Protocol: 2-byte big-endian length prefix + JSON payload.
// send() is mutex-protected so write commands can be added later without refactoring.
type Client struct {
	host   string
	port   int
	user   string
	serial string

	mu   sync.Mutex
	conn net.Conn
}

func New(host string, port int, user, serial string) *Client {
	return &Client{host: host, port: port, user: user, serial: serial}
}

// Connect establishes the TCP connection and performs the login/connect handshake.
func (c *Client) Connect(ctx context.Context) error {
	return c.connect(ctx)
}

func (c *Client) connect(ctx context.Context) error {
	if c.conn != nil {
		c.conn.Close()
	}
	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn

	resp, err := c.requestLocked(Message{"command": "login", "group": "register", "user": c.user, "pwd": ""})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if resp["status"] != "ok" {
		return fmt.Errorf("login failed: %v", resp["data"])
	}

	resp, err = c.requestLocked(Message{
		"command":       "connect",
		"group":         "register",
		"parameter":     "i-soft plus",
		"serial number": c.serial,
	})
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	if resp["status"] != "ok" {
		return fmt.Errorf("connect failed: %v", resp["data"])
	}
	slog.Info("dcm connected", "host", c.host, "serial", c.serial)
	return nil
}

// Poll sends a command and returns the response. Reconnects once on failure.
func (c *Client) Poll(ctx context.Context, group, command string) (Message, error) {
	resp, err := c.poll(group, command)
	if err != nil {
		slog.Warn("dcm poll error, reconnecting", "err", err)
		if rerr := c.connect(ctx); rerr != nil {
			return nil, fmt.Errorf("reconnect: %w", rerr)
		}
		return c.poll(group, command)
	}
	return resp, nil
}

func (c *Client) poll(group, command string) (Message, error) {
	return c.request(Message{"command": command, "group": group, "msgnumber": 1})
}

// Send sends a write command with an optional parameter (e.g. valve control).
func (c *Client) Send(ctx context.Context, group, command, parameter string) (Message, error) {
	msg := Message{"command": command, "group": group, "msgnumber": 1}
	if parameter != "" {
		msg["parameter"] = parameter
	}
	resp, err := c.request(msg)
	if err != nil {
		slog.Warn("dcm send error, reconnecting", "err", err)
		if rerr := c.connect(ctx); rerr != nil {
			return nil, fmt.Errorf("reconnect: %w", rerr)
		}
		return c.request(msg)
	}
	return resp, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// request sends a message and reads the response atomically under the mutex.
func (c *Client) request(msg Message) (Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.requestLocked(msg)
}

// requestLocked is the non-locking variant used by connect() during handshake.
func (c *Client) requestLocked(msg Message) (Message, error) {
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := Encode(c.conn, msg); err != nil {
		return nil, err
	}
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return Decode(c.conn)
}
