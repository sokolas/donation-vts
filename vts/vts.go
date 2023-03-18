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

var state string = "waiting"

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
var ParamControl chan ControlMsg = make(chan ControlMsg, 10)

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
	state = s
}

func GetState() string {
	return state
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
		internal.InfoLog.Println("reconnecting in 15 seconds...")
		setState("waiting")
		return
	}
	conn = c
	go read(c)
	if token != "" {
		setState("token_set")
	} else {
		setState("connected")
	}
}

func sendParamValue() {
	// internal.InfoLog.Printf("Sending param %v value %v", customParam, paramValue)
	msg := SetParamRequest(customParam, paramValue)
	err := conn.WriteJSON(msg)
	if err != nil {
		internal.ErrLog.Printf("error sending custom param: %v", err)
		closeAndReconnect(conn)
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

func closeAndReconnect(c *websocket.Conn) {
	closeAndReconnectIn(c, 15*time.Second)
}

func closeAndReconnectIn(c *websocket.Conn, d time.Duration) {
	setState("waiting")
	reconnectAt = time.Now().Add(d)
	c.Close()
}

func read(c *websocket.Conn) {
	for {
		msgType, message, err := c.ReadMessage()
		if err != nil {
			if state == "finished" || state == "waiting" || state == "connecting" {
				internal.InfoLog.Printf("finished/reconnecting; exit read")
				return
			}
			internal.ErrLog.Printf("read error: %v", err)
			closeAndReconnect(c)
		} else if msgType == websocket.TextMessage {
			strMsg := string(message)
			if internal.Config.LogVtsMessages {
				internal.InfoLog.Printf("received text message: %v", strMsg)
			}
			response := VTSResponse{}
			err := json.Unmarshal(message, &response)
			if err != nil {
				internal.ErrLog.Printf("can't parse reponse: %v", err)
				closeAndReconnect(c)
			} else {
				//internal.InfoLog.Printf("message type: %v", response.MessageType)
				switch response.MessageType {
				case "APIError":
					internal.ErrLog.Printf("error: %v", response.Data["message"])
					if response.Data["errorID"].(float64) == 50 {
						// user denied authentication
						internal.ErrLog.Println("authentication denied; please restart and allow plugin")
					} else if response.Data["errorID"].(float64) == 352 {
						// param already present from another plugin
						internal.ErrLog.Println("can't create custom param; please change 'customParam' in config and restart")
					}
					closeAndReconnect(c)
				case "AuthenticationTokenResponse":
					tokenStr, ok := response.Data["authenticationToken"].(string)
					if ok {
						token = tokenStr
						setState("token_set")
					} else {
						internal.ErrLog.Printf("error parsing token")
						closeAndReconnect(c)
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
							closeAndReconnect(c)
						}
					} else {
						internal.ErrLog.Printf("error parsing auth response")
						closeAndReconnect(c)
					}
				case "ParameterCreationResponse":
					setState("param_set")
				}
			}
		} else {
			internal.InfoLog.Printf("received unknown message of type %v", msgType)
		}
	}
}

func doClose() {
	if conn != nil {
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			internal.ErrLog.Println("write close: ", err)
			return
		}
	}
}

func readParam() {
	for {
		c := <-ParamControl
		if c.Msg == SetParam {
			internal.InfoLog.Printf("Setting param to %v", c.Value)
			if addParam {
				paramValue = paramValue + c.Value
			} else {
				paramValue = c.Value
			}
			if paramValue > 100 {
				paramValue = 100
			}
			if paramValue < 0 {
				paramValue = 0
			}
			decayAt = time.Now().Add(time.Duration(stayTime) * time.Second)
			if decayTime <= 1 {
				delta = paramValue
			} else {
				delta = paramValue / float64(decayTime)
			}
		}
	}
}

func Run(done chan<- int) {
	defer func() { done <- 1 }()
	go readParam()

	if stayTime >= 0 {
		decayTicker = time.NewTicker(time.Second)
		go decay()
	}

	// doConnect()
	// go read()

	for {
		select {
		case c := <-Control:
			// fmt.Printf("received control message %#v\n", c)
			switch c.Msg {
			// case SetParam:
			case Reconnect:
				if state != "waiting" && state != "finished" && state != "connecting" {
					closeAndReconnectIn(conn, 2)
				}
				UpdateConfig()
			case Done:
				setState("finished")
				if conn != nil {
					conn.Close()
				}
				return
			}

		case t := <-t.C:
			// internal.InfoLog.Println(t)
			if state == "waiting" && t.After(reconnectAt) {
				doConnect()
			}
			if state == "connected" {
				msg := TokenRequest()
				err := conn.WriteJSON(msg)
				if err != nil {
					internal.ErrLog.Printf("error sending token request: %v", err)
					closeAndReconnect(conn)
				}
				setState("token_requested")
			} else if state == "token_set" {
				msg := AuthRequest(token)
				err := conn.WriteJSON(msg)
				if err != nil {
					internal.ErrLog.Printf("error sending auth request: %v", err)
					conn.Close()
					setState("waiting")
				}
				setState("waiting_auth")
			} else if state == "authenticated" {
				msg := CreateParamRequest(customParam, internal.Config.ParamDescription)
				err := conn.WriteJSON(msg)
				if err != nil {
					internal.ErrLog.Printf("error creating custom param: %v", err)
					conn.Close()
					setState("waiting")
				}
				setState("waiting_set_param")
			} else if state == "param_set" {
				sendParamValue()
			}
		}
	}
}
