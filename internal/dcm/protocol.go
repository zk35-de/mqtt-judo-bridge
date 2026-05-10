package dcm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// Message is the generic JSON envelope the DCM uses for all commands.
type Message map[string]any

// Encode writes a framed message: 2-byte big-endian length prefix + JSON payload.
func Encode(w io.Writer, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if len(payload) > 0xFFFF {
		return fmt.Errorf("message too large: %d bytes", len(payload))
	}
	header := [2]byte{}
	binary.BigEndian.PutUint16(header[:], uint16(len(payload)))
	if _, err := w.Write(header[:]); err != nil {
		return err
	}
	_, err = w.Write(payload)
	return err
}

// Decode reads one framed message from r.
func Decode(r io.Reader) (Message, error) {
	var header [2]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint16(header[:])
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	var msg Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, err
	}
	return msg, nil
}
