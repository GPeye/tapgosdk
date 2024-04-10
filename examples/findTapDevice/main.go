package main

import (
	"github.com/GPeye/tapgosdk"
	"github.com/GPeye/tapgosdk/internal/events"
)

type onTapped struct{}

func (ot *onTapped) Handle(tap uint8) {
	println("tapdata: ", tap)
}

func main() {
	tm := tapgosdk.NewTapManager()
	onTapped := onTapped{}
	events.TappedData.Register(&onTapped)
	tm.Start()

	select {}
}
