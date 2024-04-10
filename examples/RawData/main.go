package main

import (
	"github.com/GPeye/tapgosdk"
	"github.com/GPeye/tapgosdk/events"
)

type onRaw struct{}

func (ot *onRaw) Handle(tap []byte) {
	println("tapdata: ", tap)
}

func main() {
	tm := tapgosdk.NewTapManager()
	tm.SetDefaultIputMode(tapgosdk.RawSensors)
	onRaw := onRaw{}
	events.RawData.Register(&onRaw)
	tm.Start()

	select {}
}
