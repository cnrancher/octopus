// +build test

package test

import (
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/broker"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type MemoryBroker struct {
	server  transport.Server
	backend *broker.MemoryBackend
}

func (b *MemoryBroker) Start() {
	var engine = broker.NewEngine(b.backend)
	engine.Accept(b.server)
}

func (b *MemoryBroker) Close() {
	if b.backend != nil {
		b.backend.Close(5 * time.Second)
	}
	if b.server != nil {
		_ = b.server.Close()
	}
}

func NewMemoryBroker(address string, log logr.Logger) (*MemoryBroker, error) {
	var server, err = transport.Launch(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to launch broker")
	}

	var backend = broker.NewMemoryBackend()
	backend.Logger = func(e broker.LogEvent, c *broker.Client, pkt packet.Generic, msg *packet.Message, err error) {
		if err != nil {
			log.Error(err, fmt.Sprintf("[%s]", e))
		} else if msg != nil {
			log.Info(fmt.Sprintf("[%s] %s", e, msg.String()))
		} else if pkt != nil {
			log.Info(fmt.Sprintf("[%s] %s", e, pkt.String()))
		} else {
			log.Info(fmt.Sprintf("%s", e))
		}
	}

	return &MemoryBroker{
		server:  server,
		backend: backend,
	}, nil
}
