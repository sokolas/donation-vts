package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sokolas/donation-vts/da"
	"sokolas/donation-vts/internal"
	"sokolas/donation-vts/vts"
	"time"

	"github.com/gorilla/websocket"
)

type StatusDto struct {
	VtsState string                `json:"vtsState"`
	DaState  string                `json:"daState"`
	Config   internal.AppConfigDto `json:"config"`
	AuthLink string                `json:"authLink"`
}

type ErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

var (
	//go:embed static
	res   embed.FS
	pages = map[string]string{
		"/ui": "static/index.html",
	}
	uiCalled bool = false
)

func sendStatus(done <-chan int, t *time.Ticker, c *websocket.Conn) {
	for {
		select {
		case <-done:
			return

		case <-t.C:
			status := StatusDto{
				VtsState: vts.GetState(),
				DaState:  da.GetState(),
				Config:   internal.CreateDto(),
				AuthLink: da.AuthLink,
			}
			j, _ := json.Marshal(status)
			err := c.WriteMessage(websocket.TextMessage, j)
			if err != nil {
				internal.InfoLog.Println("write:", err)
				return
			}
		}
	}
}

func setupUI() {
	// for the final build
	myfs, _ := fs.Sub(res, "static")
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.FS(myfs))))

	// for development
	// http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/api/setConfig", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("{\"ok\":false}"))
			return
		}
		req, err := io.ReadAll(r.Body)
		if err != nil {
			internal.ErrLog.Printf("Error reading setConfig request: %v", err)
			d, _ := json.Marshal(ErrorResponse{false, fmt.Sprintf("%v", err)})
			w.Write(d)
			return
		}

		dto := internal.AppConfigDto{}
		err2 := json.Unmarshal(req, &dto)
		if err2 != nil {
			internal.ErrLog.Printf("Error reading setConfig request: %v", err2)
			d, _ := json.Marshal(ErrorResponse{false, fmt.Sprintf("%v", err2)})
			w.Write(d)
			return
		}

		internal.Config = internal.MergeDto(dto)
		internal.WriteConfig()
		vts.Control <- vts.ControlMsg{Msg: vts.Reconnect}
		da.Control <- da.RECONNECT

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		status := StatusDto{
			VtsState: vts.GetState(),
			DaState:  da.GetState(),
			Config:   internal.CreateDto(),
			AuthLink: da.AuthLink,
		}
		j, _ := json.Marshal(status)
		_, err3 := w.Write(j)
		if err3 != nil {
			internal.InfoLog.Println("error writing new config:", err3)
			return
		}
	})

	upgrader := websocket.Upgrader{}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			internal.InfoLog.Print("upgrade error:", err)
			return
		}
		internal.InfoLog.Println("UI connected")
		defer c.Close()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		done := make(chan int)
		go sendStatus(done, ticker, c)

		uiCalled = true

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				internal.InfoLog.Println("read:", err)
				done <- 1
				break
			}
			internal.InfoLog.Printf("recv: %s", message)
		}
	})
}

func openBrowser(port int32) {
	if runtime.GOOS == "windows" {
		time.Sleep(2 * time.Second)
		if !uiCalled {
			cmd := exec.Command("explorer", fmt.Sprintf("http://localhost:%v/ui/", port))
			cmd.Start()
		}
	}
}

func main() {
	internal.ReadConfig()
	internal.InfoLog.Printf("Application config loaded")
	internal.DumpConfig()

	if internal.Config.AutoOpenUi {
		go openBrowser(internal.Config.DaPort)
	}

	go setupUI()

	internal.InfoLog.Println("")
	internal.InfoLog.Println("*** Press Ctrl-C to exit ***")
	internal.InfoLog.Println("")

	vts.UpdateConfig()
	vtsDone := make(chan int)
	go vts.Run(vtsDone)

	daDone := make(chan int)
	da.InitConfig()
	go da.Run(daDone)

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
