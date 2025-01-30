package mycamera

import (
	"context"
	"log"
	"radio_site/libs/myconst"
	"radio_site/libs/mystruct"
	"radio_site/libs/mywebsocket"
	"time"

	"github.com/vladimirvivien/go4vl/device"
)

func sendFrames(frames <-chan []byte) {
	for frame := range frames {
		mywebsocket.Clients.Range(func(key, value any) bool {
			key.(*mystruct.Client).FrameQueue <- frame
			return true
		})
	}
}

// try connecting camera every 5 seconds
func tryConnectCamera() {
	var camera *device.Device
	var err error
	had_err := false

	for {
		camera, err = device.Open(
			myconst.CAMERA_PATH,
			device.WithPixFormat(myconst.CAMERA_FORMAT),
		)

		if err == nil {
			break
		} else if !had_err { // only print first error to reduce clutter
			log.Println("open device:", err)
			had_err = true
		}

		time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
	}

	log.Println("Successfuly opened", camera.Name())

	had_err = false

	for {
		if err = camera.Start(context.Background()); err == nil {
			go sendFrames(camera.GetOutput())
			return
		} else if !had_err { // only print first error to reduce clutter
			log.Println("camera start:", err)
			had_err = true
		}

		time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
	}
}

func InitCamera() {
	go tryConnectCamera()
}
