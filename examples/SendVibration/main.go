package main

import (
	"time"

	"github.com/GPeye/tapgosdk"
)

type onTapped struct{}

func (ot *onTapped) Handle(tap uint8) {
	println("tapdata: ", tap)
}

func main() {
	tm := tapgosdk.NewTapManager()
	//onTapped := onTapped{}
	//events.TappedData.Register(&onTapped)
	tm.Start()

	time.Sleep(2 * time.Second)

	tm.Vibrate([]byte{50, 20, 50, 20, 50})

	select {}
}
