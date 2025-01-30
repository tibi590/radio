package myparallel

// #include "../../c_main.h"
import "C"
import (
	"errors"
	"log"
	"radio_site/libs/myconst"
)

var ErrParallelNotEnabled = errors.New("Parallel not enabled")
var ErrPortAccess = errors.New("Access denied to port")

func WritePort(pin_statuses []byte) {
    if err := CheckPerm(); err != nil {
        if err == ErrPortAccess { log.Println(err) }
        return
    }

    status_bits := C.uchar(0)

    for i, e := range pin_statuses {
        if e != '1' { continue }

        status_bits |= 1 << i
    }

    C.set_pins(status_bits)
}

func CheckPerm() (error) {
    if !myconst.USE_PARALLEL { return ErrParallelNotEnabled }
    if !C.enable_perm() { return ErrPortAccess }
    return nil
}