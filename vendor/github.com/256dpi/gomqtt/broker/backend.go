package broker

import (
	"errors"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/topic"
)

type memorySession struct {
	*session.MemorySession

	subscriptions  *topic.Tree
	storedQueue    chan *packet.Message
	temporaryQueue chan *packet.Message
	activeClient   *Client
}

func newMemorySession(backlog int) *memorySession {
	return &memorySession{
		MemorySession:  session.NewMemorySession(),
		subscriptions:  topic.NewStandardTree(),
		storedQueue:    make(chan *packet.Message, backlog),
		temporaryQueue: make(chan *packet.Message, backlog),
	}
}

func (s *memorySession) lookupSubscription(topic string) *packet.Subscription {
	// find subscription
	value := s.subscriptions.MatchFirst(topic)
	if value != nil {
		return value.(*packet.Subscription)
	}

	return nil
}

func (s *memorySession) applyQOS(msg *packet.Message) *packet.Message {
	// get subscription
	sub := s.lookupSubscription(msg.Topic)
	if sub != nil {
		// respect maximum qos
		if msg.QOS > sub.QOS {
			msg = msg.Copy()
			msg.QOS = sub.QOS
		}
	}

	return msg
}

func (s *memorySession) reuse() {
	// reset temporary queue
	s.temporaryQueue = make(chan *packet.Message, cap(s.temporaryQueue))
}

// ErrQueueFull is returned to a client that attempts two write to its own full
// queue, which would result in a deadlock.
var ErrQueueFull = errors.New("queue full")

// ErrClosing is returned to a client if the backend is closing.
var ErrClosing = errors.New("closing")

// ErrKillTimeout is returned to a client if the existing client does not close
// in time.
var ErrKillTimeout = errors.New("kill timeout")

// A MemoryBackend stores everything in memory.
type MemoryBackend struct {
	// The size of the session queue.
	SessionQueueSize int

	// The time after an error is returned while waiting on an killed existing
	// client to exit.
	KillTimeout time.Duration

	// Client configuration options. See broker.Client for details.
	ClientMaximumKeepAlive   time.Duration
	ClientParallelPublishes  int
	ClientParallelSubscribes int
	ClientInflightMessages   int
	ClientTokenTimeout       time.Duration

	// A map of username and passwords that grant read and write access.
	Credentials map[string]string

	// The Logger callback handles incoming log events.
	Logger func(LogEvent, *Client, packet.Generic, *packet.Message, error)

	activeClients     map[string]*Client
	storedSessions    map[string]*memorySession
	temporarySessions map[*Client]*memorySession
	retainedMessages  *topic.Tree
	globalMutex       sync.Mutex
	setupMutex        sync.Mutex
	closing           bool
}

// NewMemoryBackend returns a new MemoryBackend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		SessionQueueSize:  100,
		KillTimeout:       5 * time.Second,
		activeClients:     make(map[string]*Client),
		storedSessions:    make(map[string]*memorySession),
		temporarySessions: make(map[*Client]*memorySession),
		retainedMessages:  topic.NewStandardTree(),
	}
}

// Authenticate will authenticates a clients credentials.
func (m *MemoryBackend) Authenticate(_ *Client, user, password string) (bool, error) {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// return error if closing
	if m.closing {
		return false, ErrClosing
	}

	// allow all if there are no credentials
	if m.Credentials == nil {
		return true, nil
	}

	// check login
	if pw, ok := m.Credentials[user]; ok && pw == password {
		return true, nil
	}

	return false, nil
}

// Setup will close existing clients and return an appropriate session.
func (m *MemoryBackend) Setup(client *Client, id string, clean bool) (Session, bool, error) {
	// acquire setup mutex
	m.setupMutex.Lock()
	defer m.setupMutex.Unlock()

	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// return error if closing
	if m.closing {
		return nil, false, ErrClosing
	}

	// apply client settings
	client.MaximumKeepAlive = m.ClientMaximumKeepAlive
	client.ParallelPublishes = m.ClientParallelPublishes
	client.ParallelSubscribes = m.ClientParallelSubscribes
	client.InflightMessages = m.ClientInflightMessages
	client.TokenTimeout = m.ClientTokenTimeout

	// return a new temporary session if id is zero
	if len(id) == 0 {
		// create session
		sess := newMemorySession(m.SessionQueueSize)

		// set active client
		sess.activeClient = client

		// save session
		m.temporarySessions[client] = sess

		return sess, false, nil
	}

	// client id is available

	// retrieve existing client. try stored sessions before temporary sessions
	existingSession, ok := m.storedSessions[id]
	if !ok {
		if existingClient, ok2 := m.activeClients[id]; ok2 {
			existingSession, ok = m.temporarySessions[existingClient]
		}
	}

	// kill existing client if session is taken
	if ok && existingSession.activeClient != nil {
		// close client
		existingSession.activeClient.Close()

		// release global mutex to allow publish and termination, but leave the
		// setup mutex to prevent setups
		m.globalMutex.Unlock()

		// wait for client to close
		var err error
		select {
		case <-existingSession.activeClient.Closed():
			// continue
		case <-time.After(m.KillTimeout):
			err = ErrKillTimeout
		}

		// acquire mutex again
		m.globalMutex.Lock()

		// return err if set
		if err != nil {
			return nil, false, err
		}
	}

	// delete any stored session and return a temporary session if a clean
	// session is requested
	if clean {
		// delete any stored session
		delete(m.storedSessions, id)

		// create new session
		sess := newMemorySession(m.SessionQueueSize)

		// set active client
		sess.activeClient = client

		// save session
		m.temporarySessions[client] = sess

		// save client
		m.activeClients[id] = client

		return sess, false, nil
	}

	// attempt to reuse a stored session
	storedSession, ok := m.storedSessions[id]
	if ok {
		// reuse session
		storedSession.reuse()

		// set active client
		storedSession.activeClient = client

		// save client
		m.activeClients[id] = client

		return storedSession, true, nil
	}

	// otherwise create fresh session
	storedSession = newMemorySession(m.SessionQueueSize)

	// set active client
	storedSession.activeClient = client

	// save session
	m.storedSessions[id] = storedSession

	// save client
	m.activeClients[id] = client

	return storedSession, false, nil
}

// Restore is not needed at the moment.
func (m *MemoryBackend) Restore(*Client) error {
	return nil
}

// Subscribe will store the subscription and queue retained messages.
func (m *MemoryBackend) Subscribe(client *Client, subs []packet.Subscription, ack Ack) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// get session
	sess := client.Session().(*memorySession)

	// save subscription
	for _, sub := range subs {
		sess.subscriptions.Set(sub.Topic, &sub)
	}

	// call ack if provided
	if ack != nil {
		ack()
	}

	// handle all subscriptions
	for _, sub := range subs {
		// get retained messages
		values := m.retainedMessages.Search(sub.Topic)

		// publish messages
		for _, value := range values {
			// add to temporary queue or return error if queue is full
			select {
			case sess.temporaryQueue <- value.(*packet.Message):
			default:
				return ErrQueueFull
			}
		}
	}

	return nil
}

// Unsubscribe will delete the subscription.
func (m *MemoryBackend) Unsubscribe(client *Client, topics []string, ack Ack) error {
	// get session
	sess := client.Session().(*memorySession)

	// delete subscriptions
	for _, t := range topics {
		sess.subscriptions.Empty(t)
	}

	// call ack if provided
	if ack != nil {
		ack()
	}

	return nil
}

// Publish will handle retained messages and add the message to the session queues.
func (m *MemoryBackend) Publish(client *Client, msg *packet.Message, ack Ack) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// this implementation is very basic and will block the backend on every
	// publish. clients that stay connected but won't drain their queue will
	// eventually deadlock the broker. full queues are skipped if the client is
	// or is going offline

	// check retain flag
	if msg.Retain {
		if len(msg.Payload) > 0 {
			// retain message
			m.retainedMessages.Set(msg.Topic, msg.Copy())
		} else {
			// clear already retained message
			m.retainedMessages.Empty(msg.Topic)
		}
	}

	// reset retained flag
	msg.Retain = false

	// use temporary queue by default
	queue := func(s *memorySession) chan *packet.Message {
		return s.temporaryQueue
	}

	// use stored queue if qos > 0
	if msg.QOS > 0 {
		queue = func(s *memorySession) chan *packet.Message {
			return s.storedQueue
		}
	}

	// add message to temporary sessions
	for _, sess := range m.temporarySessions {
		if sub := sess.lookupSubscription(msg.Topic); sub != nil {
			if sess.activeClient == client {
				// detect deadlock when adding to own queue
				select {
				case queue(sess) <- msg:
				default:
					return ErrQueueFull
				}
			} else {
				// wait for room since client is online
				select {
				case queue(sess) <- msg:
				case <-sess.activeClient.Closing():
				}
			}
		}
	}

	// add message to stored sessions
	for _, sess := range m.storedSessions {
		if sub := sess.lookupSubscription(msg.Topic); sub != nil {
			if sess.activeClient == client {
				// detect deadlock when adding to own queue
				select {
				case queue(sess) <- msg:
				default:
					return ErrQueueFull
				}
			} else if sess.activeClient != nil {
				// wait for room since client is online
				select {
				case queue(sess) <- msg:
				case <-sess.activeClient.Closing():
				}
			} else {
				// ignore message if offline queue is full
				select {
				case queue(sess) <- msg:
				default:
				}
			}
		}
	}

	// call ack if available
	if ack != nil {
		ack()
	}

	return nil
}

// Dequeue will get the next message from the temporary or stored queue.
func (m *MemoryBackend) Dequeue(client *Client) (*packet.Message, Ack, error) {
	// mutex locking not needed

	// get session
	sess := client.Session().(*memorySession)

	// this implementation is very basic and will dequeue messages immediately
	// and not return no ack. messages are lost if the client fails to handle them

	// get next message from queue
	select {
	case msg := <-sess.temporaryQueue:
		return sess.applyQOS(msg), nil, nil
	case msg := <-sess.storedQueue:
		return sess.applyQOS(msg), nil, nil
	case <-client.Closing():
		return nil, nil, nil
	}
}

// Terminate will disassociate the session from the client.
func (m *MemoryBackend) Terminate(client *Client) error {
	// acquire global mutex
	m.globalMutex.Lock()
	defer m.globalMutex.Unlock()

	// get session
	sess := client.Session().(*memorySession)

	// release session if available
	if sess != nil {
		sess.activeClient = nil
	}

	// remove any temporary session
	delete(m.temporarySessions, client)

	// remove any saved client
	delete(m.activeClients, client.ID())

	return nil
}

// Log will call the associated logger.
func (m *MemoryBackend) Log(event LogEvent, client *Client, pkt packet.Generic, msg *packet.Message, err error) {
	// call logger if available
	if m.Logger != nil {
		m.Logger(event, client, pkt, msg, err)
	}
}

// Close will close all active clients and close the backend. The return value
// denotes if the timeout has been reached.
func (m *MemoryBackend) Close(timeout time.Duration) bool {
	// acquire global mutex
	m.globalMutex.Lock()

	// set closing
	m.closing = true

	// prepare list
	var clients []*Client

	// close temporary sessions
	for _, sess := range m.temporarySessions {
		sess.activeClient.Close()
		clients = append(clients, sess.activeClient)
	}

	// closed active stored sessions
	for _, sess := range m.storedSessions {
		if sess.activeClient != nil {
			sess.activeClient.Close()
			clients = append(clients, sess.activeClient)
		}
	}

	// release mutex to allow termination
	m.globalMutex.Unlock()

	// return early if empty
	if len(clients) == 0 {
		return true
	}

	// prepare timeout
	deadline := time.After(timeout)

	// wait for clients to close
	for _, client := range clients {
		select {
		case <-client.Closed():
			continue
		case <-deadline:
			return false
		}
	}

	return true
}
