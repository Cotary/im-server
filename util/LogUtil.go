package util

import (
	"fmt"
	"log"
)

func LogPrintf(format string, v ...interface{}) {
	fmt.Printf(format, v)
	//log.Printf(format, v)

}

func LogPrintln(v ...interface{}) {
	fmt.Println(v)
	log.Println(v)

}
