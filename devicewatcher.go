package tapgosdk

import (
	"time"

	"github.com/GPeye/tapgosdk/events"
	"tinygo.org/x/bluetooth"
)

type DeviceWatcherStatus int

const (
	Created              DeviceWatcherStatus = 0
	Started              DeviceWatcherStatus = 1
	EnumerationCompleted DeviceWatcherStatus = 2
	Stopped              DeviceWatcherStatus = 4
)

type DeviceWatcher struct {
	adapter        *bluetooth.Adapter
	timeout        time.Duration
	timer          *time.Ticker
	status         DeviceWatcherStatus
	desiredDevices int
}

func NewDeviceWatcher() *DeviceWatcher {
	d := new(DeviceWatcher)
	d.timeout = 10 * time.Second
	d.adapter = bluetooth.DefaultAdapter
	d.status = Created
	d.desiredDevices = 1
	// Enable BLE interface.
	must("enable BLE stack", d.adapter.Enable())
	return d
}

func (d *DeviceWatcher) Start() {
	d.timer = time.NewTicker(d.timeout)
	go func() {
		for {
			select {
			case <-d.timer.C:
				d.adapter.StopScan()
				d.timer.Stop()
				d.status = Stopped
				events.DeviceScanStopped.Trigger()
				println("Unable to find tap device, stopping search")
			}
		}
	}()
	err := d.adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		events.DeviceFound.Trigger(device)
	})
	if err != nil {
		panic(err.Error())
	}
}

func (d *DeviceWatcher) Stop() {
	d.adapter.StopScan()
	d.timer.Stop()
	events.DeviceScanStopped.Trigger()
	d.status = Stopped
}
