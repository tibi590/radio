package myconst

import (
	"time"

	"github.com/vladimirvivien/go4vl/v4l2"
)

const PORT = "8080"
const MAX_NUMBER_OF_PINS = 7

const PIN_FILE_PATH = "./pins.txt"

const READ_TIMEOUT = 100 * time.Second
const HEARTBEAT_TIMEOUT = 20 * time.Second

const USE_PARALLEL = true

const USE_CAMERA = true
const CAMERA_PATH = "/dev/video0"
var CAMERA_FORMAT = v4l2.PixFormat{PixelFormat: v4l2.PixelFmtH264, Width: 1280, Height: 720}