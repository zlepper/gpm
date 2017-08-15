package main

import (
	"context"
	"flag"
	"github.com/zlepper/gpm/internal"
	"log"
	"os"
	"os/signal"
)

func main() {
	var err error
	configPath := flag.String("config", "config.json", "The path to the config file. If not specified, will search in the current working directory.")

	flag.Parse()

	log.Printf("GPM version %s\n", internal.VERSION)
	pm := NewProcessManager()
	err = pm.ParseConfigFile(*configPath)
	if err != nil {
		log.Println("Could not parse config file", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- pm.StartProcesses(ctx)
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	select {
	case err = <-done:
		if err != nil {
			log.Println("Error while running processes: ", err)
		} else {
			log.Println("Processes finished by themselves.")
		}
	case <-signalChan:
		log.Println("Got interrupt, stopping processes.")
		cancel()
		select {
		case err = <-done:
			if err != nil {
				log.Println("Error while stopping processes: ", err)
			} else {
				log.Println("All processes stopped without issues.")
			}
		}
	}

}
