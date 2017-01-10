package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/wired-home/go-home/events"
	"github.com/wired-home/go-home/matrix"
	"github.com/wired-home/go-home/mqtt"
	"github.com/wired-home/go-home/worker"
)

type config struct {
	Mqtt   mqtt.Opts   `toml:"mqtt"`
	Matrix matrix.Opts `toml:"matrix"`
}

func readConfig(configFile string) config {
	_, err := os.Stat(configFile)
	if err != nil {
		log.Fatal("Config file is missing: ", configFile)
	}

	var conf config
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		log.Fatal(err)
	}
	return conf
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		os.Exit(0)
	}()

	confFile := flag.String("config", "", "Config file path")
	flag.Parse()

	conf := readConfig(*confFile)

	inbound := make(chan event.Event, 64)

	mqtt := conf.Mqtt.NewWorker(inbound)
	matrix := conf.Matrix.NewWorker(inbound)

	mqtt.Run()
	matrix.Run()

	workers := []worker.Worker{
		mqtt,
		matrix,
	}

	go func() {
		for {
			ev := <-inbound
			for _, w := range workers {
				w.Dispatch(ev)
			}
		}
	}()

	select {}
}
