package myhelper

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"
)

func InvertStatusByte(statuses []byte, pin int) {
    if statuses[pin] == '1' {
        statuses[pin] = '0'
    } else if statuses[pin] == '0' {
        statuses[pin] = '1'
    }
}

func TogglePinStatus(pin int) []byte {
    statuses := myfile.ReadPinStatuses()
    if statuses == nil { return nil }

    InvertStatusByte(statuses, pin)

    if err := myfile.WritePinFile(statuses); err != nil {
        return nil
    }
    return statuses
}

func GetData() []mystruct.Button {
    buttons := make([]mystruct.Button, myconst.MAX_NUMBER_OF_PINS)

    names := myfile.ReadPinNames()
    if names == nil { return nil }
    modes := myfile.ReadPinModes()

    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
        name := names[i]

        buttons[i] = mystruct.Button {
            Name: name,
            Num: i,
            IsToogle: modes[i] == 'T',
        }
    }

    return buttons
}
