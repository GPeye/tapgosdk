package tapgosdk

import (
	"strings"
	"time"

	"github.com/GPeye/tapgosdk/events"
	"tinygo.org/x/bluetooth"
)

type TappedEventArgs struct {
	id      string
	tapCode int
}

type TapManager struct {
	defaultInputMode         TAPInputMode
	inputModeWhenDeactivated TAPInputMode

	activated bool

	modeTimer *time.Ticker

	deviceWatcher  *DeviceWatcher
	started        bool
	restartWatcher bool

	deviceFoundHandler       deviceFoundNotifier
	deviceScanStoppedHandler deviceScanStoppedNotifier
	desiredDevices           int

	taps []TapDevice
}

func NewTapManager() *TapManager {
	tm := new(TapManager)
	tm.deviceWatcher = NewDeviceWatcher()
	tm.restartWatcher = false
	tm.deviceFoundHandler = deviceFoundNotifier{tm}
	tm.deviceScanStoppedHandler = deviceScanStoppedNotifier{tm}
	tm.desiredDevices = 1
	tm.defaultInputMode = Controller
	tm.taps = []TapDevice{}
	return tm
}

type deviceFoundNotifier struct {
	tm *TapManager
}

func (d *deviceFoundNotifier) Handle(device bluetooth.ScanResult) {
	println("found device:", device.Address.String(), device.RSSI, device.LocalName())
	println("checking if it is a Tap device")
	if strings.Contains(device.LocalName(), "Tap") && device.AdvertisementPayload.HasServiceUUID(tapService) && device.AdvertisementPayload.HasServiceUUID(nusService) {
		println("Found a Tap Device: Connecting...")
		d.tm.deviceWatcher.Stop()
		var newTapDevice = NewTapDevice(d.tm.defaultInputMode)
		newTapDevice.FromBlueToothScanResult(*d.tm.deviceWatcher.adapter, device)
		d.tm.addDevice(newTapDevice)
		//d.tm.taps = append(d.tm.taps, newTapDevice)
		println(d.tm.taps[0].fw)
		d.tm.taps[0].MakeReady()
		d.tm.taps[0].SendMode()
	} else {
		println("Not a Tap Device: Skipping...")
	}
}

type deviceScanStoppedNotifier struct {
	tm *TapManager
}

func (d deviceScanStoppedNotifier) Handle() {
	println("device scan stopped")
	if d.tm.restartWatcher {
		d.tm.restartWatcher = false
		d.tm.deviceWatcher.Start()
	}
}

func (tm *TapManager) addDevice(tapDevice TapDevice) {
	tm.taps = append(tm.taps, tapDevice)
}

func (tm *TapManager) Start() {
	if !tm.started {
		tm.started = true
		tm.activated = true
		tm.modeTimer = time.NewTicker(10 * time.Second)

		events.DeviceFound.Register(&tm.deviceFoundHandler)
		events.DeviceScanStopped.Register(&tm.deviceScanStoppedHandler)

		tm.deviceWatcher.Start()

		go func() {
			for {
				select {
				case <-tm.modeTimer.C:
					if len(tm.taps) > 0 {
						tm.taps[0].SendMode()
						//println("sending controller byte")
					}
				}
			}
		}()
	}
}

func (tm TapManager) RestartDeviceWatcher() {
	if tm.started {
		tm.restartWatcher = true
		if tm.deviceWatcher.status == Created || tm.deviceWatcher.status == Stopped {
			tm.restartWatcher = false
			println("Device Watcher Starting")
			tm.deviceWatcher.Start()
		} else if tm.deviceWatcher.status == Started {
			tm.restartWatcher = true
			tm.deviceWatcher.Stop()
		}
	}
}

func (tm *TapManager) SetDefaultIputMode(inputMode TAPInputMode) {
	tm.defaultInputMode = inputMode
}

func (tm *TapManager) Vibrate(durations []byte) {
	tm.taps[0].Vibrate(durations)
	// for i := range len(tm.taps) {
	// 	tm.taps[i].Vibrate(durations)
	// }
}
