package worker

import (
	"fmt"

	"github.com/rs/xid"
	"github.com/wired-home/go-home/events"
)

// Core struct
type Core struct {
	inbound  chan event.Event
	outbound chan event.Event
	uuid     string
}

// Worker interface
type Worker interface {
	Run()
	Dispatch(event.Event)
}

// Init - Sets up worker channels
func (c *Core) Init(inbound chan event.Event) {
	c.uuid = xid.New().String()
	c.inbound = inbound
	c.outbound = make(chan event.Event, 64)
}

// Put - Places an incoming event in inbound queue
func (c *Core) Put(ev event.Event) {
	ev.UUID = c.uuid
	fmt.Printf("[CORE-PUT:%s] %s: %s\n", ev.UUID, ev.Name, ev.Value)
	c.inbound <- ev
}

// Get - Returns an actionable event from the outbound queue
func (c *Core) Get() event.Event {
	var ev event.Event
	found := false

	for found == false {
		ev = <-c.outbound
		fmt.Printf("[CORE-GET:%s] %s: %s\n", ev.UUID, ev.Name, ev.Value)
		if ev.UUID == c.uuid {
			continue
		} else {
			found = true
		}
	}
	return ev
}

// Dispatch - Places an incoming event into the outbound queue
func (c *Core) Dispatch(ev event.Event) {
	fmt.Printf("[CORE-DISPATCH:%s] %s: %s\n", ev.UUID, ev.Name, ev.Value)
	c.outbound <- ev
}
