package matrix

import (
	"fmt"

	"github.com/matrix-org/gomatrix"
	"github.com/wired-home/go-home/events"
)

// Opts options for matrix client
type Opts struct {
	HomeServer string `toml:"home-server"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	Room       string `toml:"room"`
}

//Run - blah blah blah
func (c Opts) Run(ch <-chan events.Event) {
	creds := &gomatrix.ReqLogin{
		Type:     "m.login.password",
		User:     c.User,
		Password: c.Password,
	}

	cli, _ := gomatrix.NewClient(c.HomeServer, "", "")
	cli.Login(creds, true)

	syncer := cli.Syncer.(*gomatrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		fmt.Printf("[Matrix]: %s: %s\n", ev.Sender, ev.Content["body"])
	})

	var roomID string
	if resp, err := cli.JoinRoom(c.Room, "", nil); err != nil {
		panic(err)
	} else {
		roomID = resp.RoomID
	}

	go func() {
		for {
			event := <-ch
			msg := fmt.Sprintf("%s: %s", event.Name, event.Message)
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
