package mqtt

import (
	"fmt"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/wired-home/go-home/events"
)

// Opts options for mqtt client
type Opts struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

// Run - returns mqtt client
func (c Opts) Run(ch chan<- events.Event) {
	hostname, _ := os.Hostname()

	server := fmt.Sprintf("tcp://%s:%s", c.Host, c.Port)
	topic := "#"
	qos := 0

	connOpts := &MQTT.ClientOptions{
		ClientID:             hostname + strconv.Itoa(time.Now().Second()),
		CleanSession:         true,
		MaxReconnectInterval: 1 * time.Second,
	}

	onMessageReceived := func(client MQTT.Client, message MQTT.Message) {
		ch <- events.Event{
			Name:    message.Topic(),
			Message: string(message.Payload()),
		}

		fmt.Printf("[MQTT]: %s: %s\n", message.Topic(), message.Payload())
	}

	connOpts.AddBroker(server)
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
}
