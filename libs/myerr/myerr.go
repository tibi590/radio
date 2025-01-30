package myerr

import (
	"log"
)

func CheckErrMsg(msg string, e error) {
	if e != nil {
		log.Fatal(msg, e)
	}
}

func CheckErr(e error) {
	CheckErrMsg("", e)
}