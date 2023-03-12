package vts

import (
	"encoding/json"
	"net/http"
	"sokolas/donation-vts/internal"
	"time"

	"github.com/gorilla/websocket"
)

type ControlMsg struct {
	Msg   int32
	Value float64
}

const (
	Done      int32 = 0
	Reconnect int32 = 1
	SetParam  int32 = 2
)

// connecting, connected, reconnect, finished

var state string = "connecting"

var addr string
var token string
var customParam string
var decayTime int64 = 60
var stayTime int64 = 60
var delta float64 = 1
var addParam bool = true
var conn *websocket.Conn
var t *time.Ticker = time.NewTicker(500 * time.Millisecond)
var decayTicker *time.Ticker
var paramValue float64
var decayAt time.Time = time.Now()
var reconnectAt time.Time = time.Now()

var Control chan ControlMsg = make(chan ControlMsg, 10)

func UpdateConfig() {
	addr = internal.Config.VtsAddr
	token = internal.Config.VtsToken
	customParam = internal.Config.CustomParam
	decayTime = internal.Config.DecayTime
	if decayTime < 1 {
		decayTime = 1
	}
	stayTime = internal.Config.StayTime
	addParam = internal.Config.AddParam
}

func setState(s string) {
	internal.InfoLog.Printf("state: %v -> %v", state, s)
	if s == "connecting" {
		internal.InfoLog.Println("reconencting...")
	}
	state = s
}

func doConnect() {
	setState("connecting")
	internal.InfoLog.Println("Connecting to " + addr)
	d := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
	}
	c, _, err := d.Dial(addr, nil)
	if err != nil {
		internal.ErrLog.Println(err)
		reconnectAt = time.Now().Add(15 * time.Second)
		return
	}
	conn = c
	defer func() {
		if token != "" {
			setState("token_set")
		} else {
			setState("connected")
		}
	}()
}

func sendParamValue() {
	// internal.InfoLog.Printf("Sending param %v value %v", customParam, paramValue)
	msg := SetParamRequest(customParam, paramValue)
	err := conn.WriteJSON(msg)
	if err != nil {
		internal.ErrLog.Printf("error sending custom param: %v", err)
		conn.Close()
		setState("connecting")
	}
}

func decay() {
	for {
		<-decayTicker.C
		// internal.InfoLog.Printf("param value: %v, decay time: %v, delta: %v, decay starts at: %v\n", paramValue, decayTime, delta, decayAt)

		if stayTime >= 0 && time.Now().After(decayAt) {
			paramValue = paramValue - delta
			if paramValue < 0 {
				paramValue = 0
			}
		}
	}
}

func read() {
	for {
		if state == "connecting" {
			time.Sleep(time.Second)
		} else if state != "finished" {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if state == "finished" {
					return
				}
				internal.ErrLog.Printf("read error: %v", err)
				conn.Close()
				setState("connecting")
			} else if msgType == websocket.TextMessage {
				strMsg := string(message)
				if internal.Config.LogVtsMessages {
					internal.InfoLog.Printf("received text message: %v", strMsg)
				}
				response := VTSResponse{}
				err := json.Unmarshal(message, &response)
				if err != nil {
					internal.ErrLog.Printf("can't parse reponse: %v", err)
				} else {
					//internal.InfoLog.Printf("message type: %v", response.MessageType)
					switch response.MessageType {
					case "APIError":
						internal.ErrLog.Printf("error: %v", response.Data["message"])
						if response.Data["errorID"].(float64) == 50 {
							// user denied authentication
							internal.ErrLog.Panic("authentication denied; please restart and allow plugin")
						} else if response.Data["errorID"].(float64) == 352 {
							// param already present from another plugin
							internal.ErrLog.Panic("can't create custom param; please change 'customParam' in config.json and restart")
						}
						conn.Close()
						doConnect()
					case "AuthenticationTokenResponse":
						tokenStr, ok := response.Data["authenticationToken"].(string)
						if ok {
							token = tokenStr
							setState("token_set")
						} else {
							internal.ErrLog.Printf("error parsing token")
							//TODO reconnect?
						}
					case "AuthenticationResponse":
						authenticated, ok := response.Data["authenticated"].(bool)
						if ok {
							if authenticated {
								setState("authenticated")
								internal.Config.VtsToken = token
								internal.WriteConfig()
							} else {
								internal.ErrLog.Printf("error authenticating: %v, please approve the plugin again", response.Data["reason"])
								token = ""
								conn.Close()
								doConnect()
							}
						} else {
							internal.ErrLog.Printf("error parsing auth response")
							//TODO reconnect?
						}
					case "ParameterCreationResponse":
						setState("param_set")
					}
				}
			} else {
				internal.InfoLog.Printf("received unknown message of type %v", msgType)
			}
		} else {
			internal.ErrLog.Println("stopped reading")
			return
		}
	}
}

func doClose() {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		internal.ErrLog.Println("write close: ", err)
		return
	}
}

func Run(done chan<- int) {
	defer func() { done <- 1 }()
	if stayTime >= 0 {
		decayTicker = time.NewTicker(time.Second)
		go decay()
	}

	doConnect()
	go read()

	for {
		select {
		case c := <-Control:
			// fmt.Printf("received control message %#v\n", c)
			switch c.Msg {
			case SetParam:
				// fmt.Printf("Setting param to %v\n", c.Value)
				if addParam {
					paramValue = paramValue + c.Value
				} else {
					paramValue = c.Value
				}
				if paramValue > 100 {
					paramValue = 100
				}
				decayAt = time.Now().Add(time.Duration(stayTime) * time.Second)
				if decayTime <= 1 {
					delta = paramValue
				} else {
					delta = paramValue / float64(decayTime)
				}
			case Reconnect:
				conn.Close()
				reconnectAt = time.Now().Add(15 * time.Second)
				setState("connecting")
			case Done:
				setState("finished")
				if conn != nil {
					conn.Close()
				}
				return
			}

		case t := <-t.C:
			// internal.InfoLog.Println(t)
			if state == "connecting" && t.After(reconnectAt) {
				doConnect()
			}
			if state == "connected" {
				msg := TokenRequest()
				err := conn.WriteJSON(msg)
				if err != nil {
					internal.ErrLog.Printf("error sending token request: %v", err)
					conn.Close()
					setState("connecting")
				}
				setState("token_requested")
			} else if state == "token_set" {
				msg := AuthRequest(token)
				err := conn.WriteJSON(msg)
				if err != nil {
					internal.ErrLog.Printf("error sending auth request: %v", err)
					conn.Close()
					setState("connecting")
				}
				setState("waiting_auth")
			} else if state == "authenticated" {
				msg := CreateParamRequest(customParam, internal.Config.ParamDescription)
				err := conn.WriteJSON(msg)
				if err != nil {
					internal.ErrLog.Printf("error creating custom param: %v", err)
					conn.Close()
					setState("connecting")
				}
				setState("waiting_set_param")
			} else if state == "param_set" {
				sendParamValue()
			}
		}
	}
}
