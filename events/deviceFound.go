package events

import (
	"tinygo.org/x/bluetooth"
)

var DeviceFound deviceFound

type deviceFound struct {
	handlers []interface{ Handle(bluetooth.ScanResult) }
}

func (sr *deviceFound) Register(handler interface{ Handle(bluetooth.ScanResult) }) {
	sr.handlers = append(sr.handlers, handler)
}

func (sr deviceFound) Trigger(payload bluetooth.ScanResult) {
	for _, handler := range sr.handlers {
		go handler.Handle(payload)
	}
}
