package main

// #cgo LDFLAGS: -lm
// #include "c_main.h"
import "C"

import (
	"radio_site/libs/mycamera"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"

	"radio_site/libs/myhelper"
	"radio_site/libs/mytpl"
	"radio_site/libs/mywebsocket"

	"log"
	"net/http"
)

func page_handler(res http.ResponseWriter, req *http.Request) {
    path := req.URL.Path

    if path == "/" {
        index(res)
        return
    }

    http.NotFound(res, req)
}

func index(res http.ResponseWriter) {
    buttons := myhelper.Get_data()

    data := mystruct.IndexTemplate {
        Buttons: buttons,
        UseCamera: myconst.USE_CAMERA,
    }

    err := mytpl.Tpl.ExecuteTemplate(res, "index.html", data)

    myerr.Check_err(err)
}
    
func main() {
    if myconst.MAX_NUMBER_OF_PINS > 63 || myconst.MAX_NUMBER_OF_PINS < 1 {
        log.Fatalln("MAX_NUMBER_OF_PINS cant be bigger than 63, nor smaller than 1")
    }

    if myconst.USE_PARALLEL && !C.enable_perm() {
        log.Fatalln("Failed to get access to port!")
    }

    // if file doesnt exists, create it with default value
    myfile.Check_file()

    mytpl.Template_init()

    if myconst.USE_CAMERA {
        camera := mycamera.InitCamera()
        defer camera.Close()
    }

    http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
    http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))

    http.HandleFunc("/", page_handler)
    http.HandleFunc("/radio_ws", mywebsocket.Ws_handler)

    http.ListenAndServe(":"+myconst.PORT, nil)
}
