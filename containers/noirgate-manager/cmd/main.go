package main

import (
	"flag"
	"noirgate/api"
	"noirgate/config"
	"noirgate/core"
)

var ()

func init() {
	flag.Parse()
}

func main() {
	go core.ManageGuests()
	if *config.FlagAPI {
		go api.StartServer()
	}
	if *config.FlagTXT {
		core.RouteMessage()
	}
}
