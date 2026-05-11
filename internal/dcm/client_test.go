package dcm

import (
	"context"
	"net"
	"testing"
)

// mockServer simulates the DCM daemon over net.Pipe.
func mockServer(t *testing.T, responses []Message) (client *Client, cleanup func()) {
	t.Helper()
	serverConn, clientConn := net.Pipe()

	go func() {
		defer serverConn.Close()
		for _, resp := range responses {
			if _, err := Decode(serverConn); err != nil {
				return
			}
			if err := Encode(serverConn, resp); err != nil {
				return
			}
		}
	}()

	c := &Client{
		host:   "test",
		port:   0,
		user:   "customer",
		serial: "122907",
		conn:   clientConn,
	}
	return c, func() { clientConn.Close(); serverConn.Close() }
}

func TestLoginHandshake(t *testing.T) {
	c, cleanup := mockServer(t, []Message{
		{"command": "login", "group": "register", "status": "ok"},
		{"command": "connect", "group": "register", "status": "ok"},
	})
	defer cleanup()

	resp, err := c.request(Message{"command": "login", "group": "register", "user": "customer", "pwd": ""})
	if err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "ok" {
		t.Errorf("login status: %v", resp["status"])
	}

	resp, err = c.request(Message{"command": "connect", "group": "register", "parameter": "i-soft plus", "serial number": "122907"})
	if err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "ok" {
		t.Errorf("connect status: %v", resp["status"])
	}
}

func TestPoll(t *testing.T) {
	c, cleanup := mockServer(t, []Message{
		{"command": "salt quantity", "group": "consumption", "data": "24600", "status": "ok"},
	})
	defer cleanup()

	resp, err := c.poll("consumption", "salt quantity")
	if err != nil {
		t.Fatal(err)
	}
	if resp["data"] != "24600" {
		t.Errorf("data: got %v", resp["data"])
	}
}

func TestPollLoginError(t *testing.T) {
	c, cleanup := mockServer(t, []Message{
		{"command": "login", "status": "error", "data": "login failed"},
	})
	defer cleanup()

	// Inject a broken conn so connect() goes through mock
	serverConn2, clientConn2 := net.Pipe()
	c.conn = clientConn2
	go func() {
		defer serverConn2.Close()
		// login error
		Decode(serverConn2)
		Encode(serverConn2, Message{"status": "error", "data": "login failed"})
	}()

	err := c.connect(context.Background())
	if err == nil {
		t.Error("expected error on login failure")
	}
}
