package myfile

// #include "../../c_main.h"
import "C"
import (
	"io"
	"log"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"strconv"
	"sync"

	"os"
	"strings"
)

var (
    pinFile *os.File
    pinFileLock sync.Mutex
)

func write_port(pin_statuses []byte) {
    if !myconst.USE_PARALLEL { return }
    
    status_bits := C.uchar(0)

    for i, e := range pin_statuses {
        if e != '1' { continue }

        status_bits |= 1 << i
    }

    C.set_pins(status_bits)
}

func Write_pin_file(pin_statuses []byte) {
    write_port(pin_statuses)

    var data strings.Builder
    pin_names := Read_pin_names()

    for i := 0; i < len(pin_names); i++ {
        data.WriteString(pin_names[i])
        data.WriteByte(';')
        data.WriteByte(pin_statuses[i])
        data.WriteByte('\n')
    }

    WriteWholePinFileFD([]byte(data.String()))
}

func Read_file_lines(filepath string) [][]string {
    data := ReadWholePinFileFD()
    string_data := string(data)

    splitted_lines := strings.Split(string_data, "\n")
    lines := make([][]string, len(splitted_lines))

    for i, line := range splitted_lines{
        lines[i] = strings.Split(line, ";")
    }

    return lines[:len(lines)-1] // remove newline
}

func Read_pin_names() []string {
    lines := Read_file_lines(myconst.PIN_FILE_PATH)
    pin_names := make([]string, len(lines))

    for i, line := range lines {
        pin_names[i] = strings.TrimSpace(line[0])
    }

    return pin_names
}

func Read_pin_statuses() []byte {
    lines := Read_file_lines(myconst.PIN_FILE_PATH)
    pin_statuses := make([]byte, len(lines))

    for i, line := range lines {
        pin_statuses[i] = strings.TrimSpace(line[1])[0]
    }

    return pin_statuses
}

// ToDo use FD (FileDescription) functions instead
// more performant and safer than reopening the file
// (or store data in memory instead of the file)

func ReadPinFileFD(line int) []byte {
    pinFileLock.Lock()
    defer pinFileLock.Unlock()
    if line == -1 {
        info, err := pinFile.Stat()
        myerr.Check_err(err)
        data := make([]byte, info.Size())

        pinFile.Seek(0, 0)
        _, err = pinFile.Read(data)
        if err != io.EOF {
            myerr.Check_err(err)
        }
        return data
    }

    log.Fatalf("Not implemented!")
    return nil
}

func ReadWholePinFileFD() []byte {
    return ReadPinFileFD(-1)
}

func WritePinFileFD(data []byte, line int) {
    pinFileLock.Lock()
    defer pinFileLock.Unlock()
    if line == -1 {
        pinFile.Seek(0, 0)
        _, err := pinFile.Write(data)
        myerr.Check_err(err)
        pinFile.Sync()
        return
    }

    log.Fatalf("Not implemented!")
}

func WriteWholePinFileFD(data []byte) {
    WritePinFileFD(data, -1)
}

func Check_file() {
    text, err := os.ReadFile(myconst.PIN_FILE_PATH)
    if os.IsNotExist(err) {
        pinFile, err = os.OpenFile(myconst.PIN_FILE_PATH, os.O_RDWR | os.O_CREATE, 0644)
        myerr.Check_err(err)
        text = []byte("button 1;-")
    } else {
        myerr.Check_err(err)
        pinFile, _ = os.OpenFile(myconst.PIN_FILE_PATH, os.O_RDWR, 0644)
    }

    // if last byte isnt \n then add it
    if len(text) == 0  || text[len(text)-1] != '\n' {
        text = append(text, '\n')
        WriteWholePinFileFD(text)
    }

    lines := Read_file_lines(myconst.PIN_FILE_PATH)
    for i, line := range lines {
        if len(line) != 2 {
            lines[i] = []string{"button " + strconv.Itoa(i + 1), "-"}
        }
    }

    linesLen := len(lines)
    if linesLen == myconst.MAX_NUMBER_OF_PINS {
        return
    }

    if linesLen > myconst.MAX_NUMBER_OF_PINS {
        // remove not needed lines
        lines = lines[:myconst.MAX_NUMBER_OF_PINS]
    } else if linesLen < myconst.MAX_NUMBER_OF_PINS {
        // add "button i" lines to fill needed lines
        for i := linesLen; i < myconst.MAX_NUMBER_OF_PINS; i++ {
            lines = append(lines, []string{"button " + strconv.Itoa(i + 1), "-"})
        }
    }

    statuses := make([]byte, len(lines))

    // no need to use strings.Builder, only runs at start
    out := ""
    for i, line := range lines {
        out += line[0] + ";" + line[1] + "\n"

        if line[1] == "1" {
            statuses[i] = '1'
        } else {
            statuses[i] = '0'
        }
    }
    write_port(statuses)
    WriteWholePinFileFD([]byte(out))
}
