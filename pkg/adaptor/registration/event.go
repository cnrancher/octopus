package registration

type Event string

const (
	EventStarted   Event = "Started"
	EventStopped   Event = "Stopped"
	EventHealthy   Event = "Healthy"
	EventUnhealthy Event = "Unhealthy"
)

// Receiver is used to receive the registration event from adaptor
type Receiver interface {
	ReceiveAdaptorRegistration(adaptorName string, event Event, msg string)
}

type EventNotifier interface {
	NoticeStarted()
	NoticeStopped()
	NoticeHealthy()
	NoticeUnhealthy(msg string)
}

func NewEventNotifier(name string, receiver Receiver) EventNotifier {
	return &eventNotifier{
		name:     name,
		receiver: receiver,
	}
}

type eventNotifier struct {
	name     string
	receiver Receiver
}

func (p *eventNotifier) NoticeStarted() {
	if p.receiver == nil {
		return
	}
	p.receiver.ReceiveAdaptorRegistration(p.name, EventStarted, "")
}

func (p *eventNotifier) NoticeStopped() {
	if p.receiver == nil {
		return
	}
	p.receiver.ReceiveAdaptorRegistration(p.name, EventStopped, "")
}

func (p *eventNotifier) NoticeHealthy() {
	if p.receiver == nil {
		return
	}
	p.receiver.ReceiveAdaptorRegistration(p.name, EventHealthy, "")
}

func (p *eventNotifier) NoticeUnhealthy(msg string) {
	if p.receiver == nil {
		return
	}
	p.receiver.ReceiveAdaptorRegistration(p.name, EventUnhealthy, msg)
}
