package main

// #cgo LDFLAGS: -lm
// #include "c_main.h"
import "C"

import (
    "html/template"
    "log"
    "net/http"
    "os"
    "strconv"
)

var Tpl *template.Template

const PORT string = "8080"
const PIN_FILE_PATH string = "pin_status.txt"

func Template_init() {
    var err error

    Tpl, err = Tpl.ParseGlob("./templates/*.html")

    check_err(err)

    log.Println("Parsed templates:")
    for _, tmpl := range Tpl.Templates() {
        log.Println(" - ", tmpl.Name())
    }
}

func page_handler(res http.ResponseWriter, req *http.Request) {
    path := req.URL.Path

    if len(path) != 9 {
        index(res)
        return
    }

    if path[0:7] != "/switch" {
        return
    }

    pin, err := strconv.Atoi(path[8:])
    check_err(err)

    p := C.int(pin - 1)
    dec_data := C.int(overall_bin_status())

    status := get_pin_status(pin)
    level := C._Bool(false)

    if status == "on" {
        level = C._Bool(true)
    }

    // Placeholder. Not for actual use
    //p = p
    //dec_data = dec_data
    // level = level

    C.set_pin(p, dec_data, level)
    /*
       C.enable_perm()
       C.disable_perm()
    */

    toggle_pin_status(pin)

    http.Redirect(res, req, "/", http.StatusSeeOther)
}

func index(res http.ResponseWriter) {
    data := gen_pins()
    err := Tpl.ExecuteTemplate(res, "index.html", data)

    check_err(err)
}

func gen_pins() []Pin {
    all_pins := []Pin{}
    for i := 1; i < 8; i++ {
        status := get_pin_status(i)

        pin := Pin{
            i,
            status,
            false,
        }

        if status != "" {
            pin.IsEnabled = true
        }

        all_pins = append(all_pins, pin)
    }

    return all_pins
}

func overall_bin_status() int {
    bin_data := ""
    for i := 7; i >= 1; i-- {
        status := get_pin_status(i)

        if status == "on" {
            status = "1"
        } else {
            status = "0"
        }

        bin_data += status
    }

    dec_data, err := strconv.ParseInt(bin_data, 2, 64)
    check_err(err)

    return int(dec_data)

}

func get_pin_status(pin int) string {
    status := string(open_file(PIN_FILE_PATH)[pin-1])

    if status == "1" {
        status = "on"
    } else if status == "0" {
        status = "off"
    } else if status == "-" {
        status = ""
    }

    return status
}

func toggle_pin_status(pin int) {
    data := string(open_file(PIN_FILE_PATH))
    altered_data := ""

    for i := 0; i < len(data); i++ {
        if i == pin-1 {
            if string(data[i]) == "1" {
                altered_data += "0"
            } else {
                altered_data += "1"
            }
        } else {
            altered_data += string(data[i])
        }
    }

    altered_data += "\n"

    write_file(PIN_FILE_PATH, []byte(altered_data))
}

func write_file(filename string, data []byte) {
    err := os.WriteFile(filename, data, 0644)

    check_err(err)
}

func open_file(filename string) []byte {
    data, err := os.ReadFile(filename)

    check_err(err)

    return data[:len(data)-1]
}

func check_err(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

func main() {
    fs_css := http.FileServer(http.Dir("./css"))
    http.Handle("/css/", http.StripPrefix("/css", fs_css))

    http.HandleFunc("/", page_handler)

    Template_init()

    // if file doesnt exists, create it with default value
    _, err := os.Stat(PIN_FILE_PATH)
    if os.IsNotExist(err) {
        write_file(PIN_FILE_PATH, []byte("0000----\n"))
    } else {
        check_err(err)
    }

    open_file(PIN_FILE_PATH)

    http.ListenAndServe(":"+PORT, nil)
}

type Pin struct {
    Num       int
    Status    string
    IsEnabled bool
}
