package tapgosdk

import (
	"github.com/GPeye/tapgosdk/internal/events"
	"tinygo.org/x/bluetooth"
)

type TapProperties struct {
	tapData   bluetooth.DeviceCharacteristic
	mouseData bluetooth.DeviceCharacteristic
	nusRx     bluetooth.DeviceCharacteristic
	fwVersion int
}

type TapDevice struct {
	tapDataValueChangedAssigned   bool
	mouseDataValueChangedAssigned bool

	adapter   *bluetooth.Adapter
	device    bluetooth.Device
	rx        *bluetooth.DeviceCharacteristic
	tapData   *bluetooth.DeviceCharacteristic
	mousedata bluetooth.DeviceCharacteristic
	fwChar    *bluetooth.DeviceCharacteristic

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

	chars, err := tapServ.DiscoverCharacteristics([]bluetooth.UUID{tapDataCharacteristic})
	if err != nil {
		println(err)
	}
	if len(chars) == 0 {
		panic("could not find tap data characteristic")
	}
	td.tapData = &chars[0]
	println("found Tap Data characteristic", td.tapData.UUID().String())

	chars, err = nusServ.DiscoverCharacteristics([]bluetooth.UUID{tapModeCharacteristic})
	if err != nil {
		println(err)
	}
	if len(chars) == 0 {
		panic("could not find tap data characteristic")
	}
	td.rx = &chars[0]
	println("found Tap Mode characteristic", td.rx.UUID().String())

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

func (td *TapDevice) GetTapProperties() TapProperties {
	return TapProperties{}
}

func (td *TapDevice) MakeReady() {
	if td.isConnected {
		td.tapData.EnableNotifications(func(buf []byte) {
			events.TappedData.Trigger(uint8(buf[0]))
		})
		td.isReady = true
	}
}

func (td *TapDevice) SendMode() {
	println("Sending Mode: ", byte(Controller))
	if td.isReady && td.isConnected {
		data := []byte{0x3, 0xc, 0x0, byte(Controller)}
		td.rx.WriteWithoutResponse(data)
	}
}
