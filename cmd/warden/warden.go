package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

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
	w.Run()
}
