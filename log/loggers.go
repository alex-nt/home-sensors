package log

import (
	"log"
	"os"
)

var (
	ErrorLog *log.Logger
	InfoLog  *log.Logger
)

func init() {
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}