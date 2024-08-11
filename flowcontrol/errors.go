package flowcontrol

import "errors"

var (
	ErrMetadataNotFound             = errors.New("metadata not found")
	ErrHardcodedConfigKeysMissmatch = errors.New("local key: " + noexport_CONFIG_KEY + " is not the same as key in flowcore")
	ErrClientsideError              = errors.New("clientside error")
	ErrClientsideUnimplemented      = errors.New("clientside unimplemented")
	ErrUnknownReply                 = errors.New("unknown reply")
	ErrDeadClient                   = errors.New("client response taken too long, client treated as a dead")
)
