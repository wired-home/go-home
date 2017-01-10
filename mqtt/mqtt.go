package mqtt

import (
	"fmt"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/wired-home/go-home/events"
	"github.com/wired-home/go-home/worker"
)

// Opts for mqtt client from config file
type Opts struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

// Worker - MQTT worker
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

// Run - returns mqtt client
func (w Worker) Run() {
	hostname, _ := os.Hostname()
	server := fmt.Sprintf("tcp://%s:%s", w.opts.Host, w.opts.Port)
	clientid := hostname + strconv.Itoa(time.Now().Second())
	topic := "#"
	qos := 0

	connOpts := MQTT.NewClientOptions().
		AddBroker(server).
		SetClientID(clientid).
		SetCleanSession(true)

	onMessageReceived := func(client MQTT.Client, message MQTT.Message) {
		w.Put(
			event.Event{
				Name:  message.Topic(),
				Value: string(message.Payload()),
			},
		)
	}

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, byte(qos), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", server)
	}

	go func() {
		for {
			event := w.Get()
			if token := client.Publish(event.Name, byte(qos), true, event.Value); token.Wait() && token.Error() != nil {
				fmt.Println(token.Error())
			}
		}
	}()

}
