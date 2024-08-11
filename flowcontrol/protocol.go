package flowcontrol

import "time"

// # Flow Controller Protocol.
//
// Protocol consists of 2 stages:
//
// 1. Flow Controller sends primary messages-signals:
//
//   - [GracefulShutdown]
//   - [MetadataUpdated]
//   - [WaitFor]
//   - [Continue]
//   - [Status]
//
// 2. Flow Controller Messages:

const PROTOCOL = -0

// =====================================================

// ====================== Stage 1 ======================

// Primary messages-signals from controller
const (
	GracefulShutdown MessageFrom = iota

	// # Guranteed to be send only if [ServiceMetadata] changed
	MetadataUpdated

	// # Guranteed to be followed by [Continue]
	WaitFor
	// # Guranteed to be send after  [WaitFor]
	Continue

	// # Guranteed to be send continiously if client replies with [Pending]
	//
	// Will be sent with every [STATUS_SPAM_TIMEOUT]
	Status
)
const STATUS_SPAM_TIMEOUT = 100 * time.Millisecond

// ====================== Stage 2 ======================

// Non-error client-side replies
const (
	OK MessageTo = iota + 200

	// # Protocol Guarantees
	//
	// If client replies with [Pending]
	//
	// Then controller will send [Status] every [STATUS_SPAM_TIMEOUT]
	//
	// Untill [ServiceManager] replies with Non-[Pending] message
	Pending
)

// Error client-side replies
const (
	// # Client-side error
	Error MessageTo = iota + 400
	// # Default reply on unimplemented stage 1 message
	Unimplemented
)
