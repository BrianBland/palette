package main

import (
	"log"
	"os"

	"github.com/BrianBland/go-hue"
	"github.com/BrianBland/palette"
	"github.com/BrianBland/palette/server"
)

func main() {
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	bridges, err := hue.FindBridgesUsingCloud()
	if err != nil {
		log.Fatal("Failed to find bridge:", err)
	}
	if len(bridges) == 0 {
		log.Fatal("No bridges found")
	}
	log.Print("Found bridges:", bridges)
	bridge := bridges[0]

	p, err := palette.LoadFromConfig(bridge)
	if err != nil {
		log.Print("Failed to load config, making new user. Error:", err)
		p, err = palette.New(bridge)
		if err != nil {
			log.Fatal("Failed to create new config:", err)
		}
		err = p.SaveToConfig()
		if err != nil {
			log.Fatal("Failed to save config:", err)
		}
	}

	s := server.New(p)
	log.Fatal(s.ListenAndServe(addr))
}
