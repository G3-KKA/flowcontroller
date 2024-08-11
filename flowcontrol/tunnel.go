package flowcontrol

// # High Level API for Tunnel Stage 2
//
// Available [clientsideReply] listed in Stage 2 of flowcontrol/protocol.go
type ReplyFunc func(clientsideReply MessageTo)

// # Message that controller sends to registered service
//
// May be a primary message or reply on a previous [MessageTo]
type MessageFrom int

// # Message that service sends to controller
type MessageTo int

// # Pair of channels connecting controller and registred service
type ControlTunnel struct {
	tun tunnel
}

// # Flow Controller Messages:
//
//   - [GracefulShutdown]
//   - [MetadataUpdated]
//   - [WaitFor] + [Continue]
//   - [Status]
func (ctunnel *ControlTunnel) ReadMessage() (MessageFrom, ReplyFunc) {
	f := ReplyFunc(func(msg MessageTo) {
		ctunnel.tun.to <- msg
	})
	return <-ctunnel.tun.from, f
}

// # Warning! Low Level API, be careful.
//
// Use to gain access to channels behind controller tunnel api
func (ctunnel *ControlTunnel) LowLevelAPI() tunnel {
	return ctunnel.tun
}

// Flow Controller Protocol guarantees that per every message send to Flow Controller - there gonna be a response.
// That response must be read by the client
type tunnel struct {
	from chan MessageFrom
	to   chan MessageTo
}

// # Warning! Low Level API, be careful.
//
// Use for stage 1 and 3, see flowcontrol/protocol.go for details
func (tun *tunnel) Read() <-chan MessageFrom {
	return tun.from
}

// # Warning! Low Level API, be careful.
//
// Use for stage 2, see flowcontrol/protocol.go for details
func (tun *tunnel) Write() chan<- MessageTo {
	return tun.to
}
