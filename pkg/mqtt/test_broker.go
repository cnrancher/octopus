// +build test

package mqtt

import (
	"fmt"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/broker"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type TestMemoryBroker struct {
	server  transport.Server
	backend *broker.MemoryBackend
}

func (b *TestMemoryBroker) Start() {
	var engine = broker.NewEngine(b.backend)
	engine.Accept(b.server)
}

func (b *TestMemoryBroker) Close() {
	if b.backend != nil {
		b.backend.Close(5 * time.Second)
	}
	if b.server != nil {
		_ = b.server.Close()
	}
}

func NewTestMemoryBroker(address string, log logr.Logger) (*TestMemoryBroker, error) {
	var server, err = transport.Launch(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to launch broker")
	}

	var backend = broker.NewMemoryBackend()
	backend.Logger = func(e broker.LogEvent, c *broker.Client, pkt packet.Generic, msg *packet.Message, err error) {
		if err != nil {
			log.Info("[%s] %s", e, err)
		} else if msg != nil {
			log.Info("[%s] %s", e, msg.String())
		} else if pkt != nil {
			log.Info("[%s] %s", e, pkt.String())
		} else {
			log.Info("%s", e)
		}
	}

	return &TestMemoryBroker{
		server:  server,
		backend: backend,
	}, nil
}

type testingTLogger struct {
	t *testing.T
}

func (t *testingTLogger) Info(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) != 0 {
		t.t.Logf(msg, keysAndValues...)
	} else {
		t.t.Log(msg)
	}
}

func (t *testingTLogger) Enabled() bool {
	return true
}

func (t *testingTLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) != 0 {
		t.t.Errorf("%s: %v", fmt.Sprintf(msg, keysAndValues...), err)
	} else {
		t.t.Errorf("%s: %v", msg, err)
	}
}

func (t *testingTLogger) V(level int) logr.InfoLogger {
	return t
}

func (t *testingTLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return t
}

func (t *testingTLogger) WithName(name string) logr.Logger {
	return t
}
