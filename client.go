package client

import (
	"github.com/yam8511/go-pomelo-client/codec"
)

// Callback represents the callback type which will be called
// when the correspond events is occurred.
type Callback func(data []byte)

// NewConnector create a new Connector
func NewConnector() *Connector {
	return &Connector{
		die:       make(chan byte),
		codec:     codec.NewDecoder(),
		chSend:    make(chan []byte, 64),
		mid:       1,
		events:    map[string]Callback{},
		responses: map[uint]Callback{},
	}
}
