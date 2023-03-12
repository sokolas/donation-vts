package main

import (
	"fmt"
	"os"
	"os/signal"
	"sokolas/donation-vts/da"
	"sokolas/donation-vts/internal"
	"sokolas/donation-vts/vts"
	/*
		"net/url"
		"os"
		"github.com/gorilla/websocket"*/)

//var vtsAddr = flag.String("vtsAddr", )

func main() {
	internal.ReadConfig()
	internal.InfoLog.Printf("Application config loaded")
	internal.DumpConfig()

	internal.InfoLog.Println("\n*** Press Ctrl-C to exit ***\n")

	vts.UpdateConfig()
	vtsDone := make(chan int)
	go vts.Run(vtsDone)

	daDone := make(chan int)
	da.UpdateConfig()
	go da.Run(daDone)

	// vts.Control <- vts.ControlMsg{Msg: vts.SetParam, Value: 50}
	// vts.Control <- vts.ControlMsg{Msg: vts.SetParam, Value: 50}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for {
		select {
		case <-vtsDone:
			return
		case <-daDone:
			return
		case <-interrupt:
			fmt.Println("interrupted")
			vts.Control <- vts.ControlMsg{Msg: vts.Done, Value: 0}
			da.Control <- da.DONE
		}
	}

}
