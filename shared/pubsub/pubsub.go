package pubsub

type Subject struct {
	// subscribers is all subject subscribers.
	Subscribers map[string]func(any)
}

type PubSub struct {
	// subjects is all pubsub subjects.
	subjects map[string]*Subject
}

func New() *PubSub {
	return &PubSub{
		subjects: make(map[string]*Subject),
	}
}

// Subscribe subscribe to a subject.
func (p *PubSub) Subscribe(subject string, key string, calback func(any)) {
	s, ok := p.subjects[subject]
	if !ok {
		s = p.createSubject(subject)
	}
	s.Subscribers[key] = calback
}

// Unsubscribe unsubscribes from a subject.
func (p *PubSub) Unsubscribe(subject string, key string) {
	delete(p.subjects, key)
}

// Publish publish data to a subject.
func (p *PubSub) Publish(subject string, d any) {
	s, ok := p.subjects[subject]
	if !ok {
		s = p.createSubject(subject)
	}
	for _, v := range s.Subscribers {
		v(d)
	}
}

// createSubject creates a new subject.
func (p *PubSub) createSubject(subject string) *Subject {
	s := &Subject{
		Subscribers: make(map[string]func(any)),
	}
	p.subjects[subject] = s
	return s
}
