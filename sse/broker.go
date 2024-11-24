package sse

type Broker struct {
	// users is a map where the key is the user id
	// and the value is a slice of channels to connections
	// for that user id
	users map[string][]chan Event

	// actions is a channel of functions to call
	// in the broker's goroutine. The broker executes
	// everything in that single goroutine to avoid
	// data races.
	actions chan func()
}

// Run executes in a goroutine. It simply gets and
// calls functions.
func (b *Broker) Run() {
	for a := range b.actions {
		a()
	}
}

func NewBroker() *Broker {
	b := &Broker{
		users:   make(map[string][]chan Event),
		actions: make(chan func()),
	}
	go b.Run()
	return b
}

// AddUserChan adds a channel for user with given id.
func (b *Broker) AddUserChan(id string, ch chan Event) {
	b.actions <- func() {
		b.users[id] = append(b.users[id], ch)
	}
}

// RemoveUserchan removes a channel for a user with the given id.
func (b *Broker) RemoveUserChan(id string, ch chan Event) {
	// The broker may be trying to send to
	// ch, but nothing is receiving. Pump ch
	// to prevent broker from getting stuck.
	go func() {
		for range ch {
		}
	}()

	b.actions <- func() {
		chs := b.users[id]
		i := 0
		for _, c := range chs {
			if c != ch {
				chs[i] = c
				i = i + 1
			}
		}
		if i == 0 {
			delete(b.users, id)
		} else {
			b.users[id] = chs[:i]
		}
		// Close channel to break loop at beginning
		// of removeUserChan.
		// This must be done in broker goroutine
		// to ensure that broker does not send to
		// closed goroutine.
		close(ch)
	}
}

// Broadcast sends a message to all users.
func (b *Broker) BroadcastEvent(event Event) {
	b.actions <- func() {
		for _, chs := range b.users {
			for _, ch := range chs {
				ch <- event
			}
		}
	}
}
