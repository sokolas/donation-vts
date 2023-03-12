package internal

import (
	"io"
	"log"
	"os"
)

var logfile, _ = os.Create("log.txt")

var infoWriter = io.MultiWriter(logfile, os.Stdout)
var warnWriter = io.MultiWriter(logfile, os.Stdout)
var errWriter = io.MultiWriter(logfile, os.Stderr)

var InfoLog *log.Logger = log.New(infoWriter, "INFO  ", log.Ldate|log.Ltime|log.Lshortfile)
var WarnLog *log.Logger = log.New(warnWriter, "WARN  ", log.Ldate|log.Ltime|log.Lshortfile)
var ErrLog *log.Logger = log.New(errWriter, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
