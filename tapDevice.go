package tapgosdk

import (
	"encoding/binary"
	"reflect"

	"github.com/GPeye/tapgosdk/events"
	"tinygo.org/x/bluetooth"
)

type TapDevice struct {
	tapDataValueChangedAssigned   bool
	mouseDataValueChangedAssigned bool

	adapter    *bluetooth.Adapter
	device     bluetooth.Device
	rx         *bluetooth.DeviceCharacteristic
	tx         *bluetooth.DeviceCharacteristic
	tapData    *bluetooth.DeviceCharacteristic
	mousedata  bluetooth.DeviceCharacteristic
	uiCommands *bluetooth.DeviceCharacteristic
	fwChar     *bluetooth.DeviceCharacteristic

	isReady       bool
	isConnected   bool
	fw            int
	inputMode     TAPInputMode
	supportsMouse bool
}

func NewTapDevice(mode TAPInputMode) TapDevice {
	td := new(TapDevice)
	td.fw = 0
	td.isReady = false
	td.isConnected = false
	td.inputMode = mode
	td.supportsMouse = false
	td.tapDataValueChangedAssigned = false
	td.mouseDataValueChangedAssigned = false

	return *td
}

func (td TapDevice) InputMode() TAPInputMode {
	return td.inputMode
}

func (td TapDevice) IsReady() bool {
	return td.isReady
}

func (td TapDevice) SupportsMouse() bool {
	return td.supportsMouse
}

func (td TapDevice) Identifier() bluetooth.MACAddress {
	return td.device.Address.MACAddress
}

func (td TapDevice) Name() string {
	return "name"
}

func (td TapDevice) FW() int {
	return td.fw
}

func (td *TapDevice) FromBlueToothScanResult(adapter bluetooth.Adapter, device bluetooth.ScanResult) TapDevice {
	td.adapter = &adapter
	var err error = nil
	td.device, err = adapter.Connect(device.Address, bluetooth.ConnectionParams{})
	if err != nil {
		panic(err.Error())
	}

	td.isConnected = true

	println("discovering services/characteristics")
	srvcs, err := td.device.DiscoverServices([]bluetooth.UUID{tapService, nusService, deviceInfoService})
	must("discover services", err)

	if len(srvcs) < 3 {
		panic("Could not find all Tap Services")
	}

	tapServ := srvcs[0]
	nusServ := srvcs[1]
	infoServ := srvcs[2]

	println("found Tap Service", tapServ.UUID().String())
	println("found Nus Service", nusServ.UUID().String())
	println("found Info Service", infoServ.UUID().String())

	chars, err := tapServ.DiscoverCharacteristics([]bluetooth.UUID{tapDataCharacteristic, uiCmdCharacteristic})
	if err != nil {
		println(err)
	}
	if len(chars) < 2 {
		panic("could not find tap data characteristic")
	}
	td.tapData = &chars[0]
	td.uiCommands = &chars[1]
	println("found Tap Data characteristic", td.tapData.UUID().String())
	println("found Tap UI Command characteristic", td.uiCommands.UUID().String())

	chars, err = nusServ.DiscoverCharacteristics([]bluetooth.UUID{tapModeCharacteristic, rawSensorsCharacteristic})
	if err != nil {
		println(err)
	}
	if len(chars) < 2 {
		panic("could not find tap data characteristic")
	}
	td.rx = &chars[0]
	td.tx = &chars[1]
	println("found Tap Mode characteristic", td.rx.UUID().String())
	println("found Raw Mode characteristic", td.tx.UUID().String())

	chars, err = infoServ.DiscoverCharacteristics([]bluetooth.UUID{fwVersionCharacteristic})
	if err != nil {
		println(err)
	}
	if len(chars) == 0 {
		panic("could not find tap data characteristic")
	}
	td.fwChar = &chars[0]
	println("found Fw Version characteristic", td.fwChar.UUID().String())
	var fwdata []byte = []byte{0}
	_, err = td.fwChar.Read(fwdata)
	if err != nil {
		println(err)
	}

	println(int(fwdata[0]))
	td.fw = int(fwdata[0])

	return TapDevice{}
}

func (td *TapDevice) Vibrate(durations []byte) {
	println("Sending Vibrate")
	if td.uiCommands == nil || !td.isReady {
		return
	}
	//byteDurations := IntsToBytesBE(durations)
	//println(byteDurations)
	var data [20]byte
	data[0] = 0
	data[1] = 2
	for i := range len(data) - 2 {
		if i < len(durations) {
			data[i+2] = durations[i]
		} else {
			data[i+2] = 0
		}
	}
	td.uiCommands.WriteWithoutResponse(data[:])
}

func IntsToBytesBE(i []int) []byte {
	intSize := int(reflect.TypeOf(i).Elem().Size())
	b := make([]byte, intSize*len(i))
	for n, s := range i {
		switch intSize {
		case 64 / 8:
			binary.BigEndian.PutUint64(b[intSize*n:], uint64(s))
		case 32 / 8:
			binary.BigEndian.PutUint32(b[intSize*n:], uint32(s))
		default:
			panic("unreachable")
		}
	}
	return b
}

func (td *TapDevice) MakeReady() {
	if td.isConnected {
		switch td.inputMode {
		case Controller:
			td.tapData.EnableNotifications(func(buf []byte) {
				events.TappedData.Trigger(uint8(buf[0]))
			})
		case RawSensors:
			err := td.tx.EnableNotifications(func(buf []byte) {
				println(buf)
				events.RawData.Trigger(buf)
			})
			if err != nil {
				println(err)
			}
		default:
		}
		td.isReady = true
	}
}

func (td *TapDevice) SendMode() {
	println("Sending Mode: ", byte(td.inputMode))
	if td.isReady && td.isConnected {
		var data []byte
		if td.inputMode == RawSensors {
			data = []byte{0x3, 0xc, 0x0, byte(td.inputMode), 0x1, 0x2, 0x1}
		} else {
			data = []byte{0x3, 0xc, 0x0, byte(td.inputMode)}
		}
		td.rx.WriteWithoutResponse(data)
	}
}
