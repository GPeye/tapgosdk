package tapgosdk

type TAPInputMode int

const (
	Controller             TAPInputMode = 0x1
	ControllerWithMouseHID TAPInputMode = 0x03
	RawSensors             TAPInputMode = 0xa
	Text                   TAPInputMode = 0x0
	Null                   TAPInputMode = 0xff
)
