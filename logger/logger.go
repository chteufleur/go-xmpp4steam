package logger

import (
	"io"
	"log"
)

var (
	Info  *log.Logger
	Debug *log.Logger
	Error *log.Logger
)

func Init(infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Info = log.New(infoHandle, "INFO  : ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(warningHandle, "DEBUG : ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR : ", log.Ldate|log.Ltime|log.Lshortfile)
}
