package dcm

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	msg := Message{"command": "login", "group": "register", "user": "customer", "pwd": ""}
	var buf bytes.Buffer
	if err := Encode(&buf, msg); err != nil {
		t.Fatal(err)
	}
	// first 2 bytes are length
	if buf.Len() < 2 {
		t.Fatal("encoded too short")
	}
	got, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if got["command"] != "login" {
		t.Errorf("command: got %v", got["command"])
	}
}

func TestEncodeDecodeLargeMessage(t *testing.T) {
	// weekly data with spaces, like real device responses
	msg := Message{"command": "water weekly", "group": "consumption", "data": " -1 10 195 292 305 107 21", "status": "ok"}
	var buf bytes.Buffer
	if err := Encode(&buf, msg); err != nil {
		t.Fatal(err)
	}
	got, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if got["data"] != " -1 10 195 292 305 107 21" {
		t.Errorf("data: got %v", got["data"])
	}
}

func TestEncodeTooBig(t *testing.T) {
	big := make([]byte, 0xFFFF+1)
	msg := Message{"x": string(big)}
	var buf bytes.Buffer
	if err := Encode(&buf, msg); err == nil {
		t.Error("expected error for oversized message")
	}
}
