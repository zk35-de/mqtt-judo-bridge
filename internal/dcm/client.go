package dcm

// Client connects to the Judo DevCommManagerDaemon on TCP port 8833.
// Protocol: 2-byte big-endian length prefix + JSON payload.
// After login handshake the daemon pushes notifications for subscribed groups.
type Client struct{}
