package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BrianBland/warden"
)

func main() {
	configFile := "warden.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	var config warden.Config
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("Failed to read config file:", err)
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatalln("Failed to parse config file:", err)
	}
	w, err := warden.New(config)
	if err != nil {
		log.Fatalln("Failed to create warden:", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signalChan {
			log.Println("Received an interrupt, stopping warden...")
			w.Cleanup()
			os.Exit(1)
		}
	}()

	w.Run()
}
