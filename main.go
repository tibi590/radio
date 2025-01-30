package main

// #cgo LDFLAGS: -lm
// #include "c_main.h"
import "C"

import (
	"radio_site/libs/mycamera"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/myfile"
	"radio_site/libs/myparallel"
	"radio_site/libs/mystruct"

	"radio_site/libs/myhelper"
	"radio_site/libs/mytpl"
	"radio_site/libs/mywebsocket"

	"log"
	"net/http"
)

func pageHandler(res http.ResponseWriter, req *http.Request) {
    path := req.URL.Path

    if path == "/" {
        index(res)
        return
    }

    http.NotFound(res, req)
}

func index(res http.ResponseWriter) {
    buttons := myhelper.GetData()
    if buttons == nil {
        http.Error(res, "Failed to read pins file", 500)
        return
    }

    data := mystruct.IndexTemplate {
        Buttons: buttons,
    }

    err := mytpl.Tpl.ExecuteTemplate(res, "index.html", data)

    myerr.CheckErr(err)
}
    
func main() {
    if myconst.MAX_NUMBER_OF_PINS > 63 || myconst.MAX_NUMBER_OF_PINS < 1 {
        log.Fatalln("MAX_NUMBER_OF_PINS cant be bigger than 63, nor smaller than 1")
    }

    if err := myparallel.CheckPerm(); err == myparallel.ErrPortAccess {
        log.Fatalln(err)
    }

    // if file doesnt exists, create it with default value
    myfile.CheckFile()

    mytpl.TemplateInit()

    if myconst.USE_CAMERA {
        mycamera.InitCamera()
    }

    http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
    http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))

    http.HandleFunc("/", pageHandler)
    http.HandleFunc("/radio_ws", mywebsocket.WsHandler)

    http.ListenAndServe(":"+myconst.PORT, nil)
}
