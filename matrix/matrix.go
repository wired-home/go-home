package matrix

import (
	"fmt"
	"strings"

	"github.com/matrix-org/gomatrix"

	"github.com/wired-home/go-home/events"
	"github.com/wired-home/go-home/worker"
)

// Opts options for matrix client
type Opts struct {
	HomeServer string `toml:"home-server"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	Room       string `toml:"room"`
}

// Worker - Matrix worker
type Worker struct {
	worker.Core
	opts Opts
}

// NewWorker - Create a new worker
func (c Opts) NewWorker(inbound chan event.Event) *Worker {
	worker := Worker{
		opts: c,
	}
	worker.Init(inbound)
	return &worker
}

//Run - blah blah blah
func (w Worker) Run() {
	creds := &gomatrix.ReqLogin{
		Type:     "m.login.password",
		User:     w.opts.User,
		Password: w.opts.Password,
	}

	cli, _ := gomatrix.NewClient(w.opts.HomeServer, "", "")
	cli.Login(creds, true)

	syncer := cli.Syncer.(*gomatrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Sender != cli.UserID {
			fmt.Printf("[Matrix]: %s: %s\n", ev.Sender, ev.Content["body"])
			split := strings.Split(ev.Content["body"].(string), ":")
			name, value := split[0], split[1]
			w.Put(
				event.Event{
					Name:  name,
					Value: value,
				},
			)
		}
	})

	var roomID string
	if resp, err := cli.JoinRoom(w.opts.Room, "", nil); err != nil {
		panic(err)
	} else {
		roomID = resp.RoomID
	}

	go func() {
		for {
			event := w.Get()
			msg := fmt.Sprintf("%s:%s", event.Name, event.Value)
			cli.SendText(roomID, msg)
		}
	}()

	go func() {
		for {
			if err := cli.Sync(); err != nil {
				fmt.Println("Sync() returned ", err)
			}
		}
	}()
}
