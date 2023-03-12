package da

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sokolas/donation-vts/internal"
	"sokolas/donation-vts/vts"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DONE = 1

	AUTH_BASE_URL = "https://www.donationalerts.com/oauth/authorize"
	USER_BASE_URL = "https://www.donationalerts.com/api/v1/user/oauth"
	SUBSCRIBE_URL = "https://www.donationalerts.com/api/v1/centrifuge/subscribe"
	WEBSOCKET_URL = "wss://centrifugo.donationalerts.com/connection/websocket"

	SCOPES = "oauth-donation-subscribe oauth-user-show"

	REDIRECT_RESPONSE = `<!DOCTYPE html>
	<html>
	<body>
	<script>
	let hash = location.hash ? location.hash : ""
	location.href= "/r" + hash.replace('#', '?')
	</script>
	</body>
	</html>`

	SUCCESS_RESPONSE = `<!DOCTYPE html>
	<html>
	<body>
	<H2>Success! You can close the window</H2>
	</body>
	</html>`

	ERROR_RESPONSE = `<!DOCTYPE html>
	<html>
	<body>
	<H2>Error: %v! Please try again</H2>
	</body>
	</html>`
)

var Control chan int = make(chan int, 10)

var state string = "connecting"

var conn *websocket.Conn

var accessToken string
var appId string
var clientKey string
var userId int64
var listenPort int32
var socketConnectionToken string
var channelToken string
var multiplier float64 = 1

var msgId int64 = 0

var t = time.NewTicker(time.Second)
var reconnectAt time.Time = time.Now()

func UpdateConfig() {
	accessToken = internal.Config.DaToken
	listenPort = internal.Config.DaPort
	appId = internal.Config.DaAppId
	multiplier = internal.Config.Multiplier
}

func setState(s string) {
	internal.InfoLog.Printf("state: %v -> %v", state, s)
	state = s
}

func nextId() int64 {
	msgId++
	return msgId
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	// internal.InfoLog.Println(req.RequestURI)
	u, _ := url.Parse(req.RequestURI)
	if u.Path == "/" {
		fmt.Fprint(w, REDIRECT_RESPONSE)
		return
	}
	if u.Path != "/r" {
		return
	}

	if state != "connecting" {
		return
	}

	p := u.Query()
	// internal.InfoLog.Println(p)
	if p.Has("error") {
		internal.ErrLog.Printf("Error from donationalerts: %v", p.Get("error"))
		fmt.Fprintf(w, ERROR_RESPONSE, p.Get("error"))
	} else if p.Has("access_token") {
		accessToken = p.Get("access_token")
		fmt.Fprintf(w, SUCCESS_RESPONSE)
		internal.Config.DaToken = accessToken
		internal.WriteConfig()
		setState("access_token_set")
	}
}

func Prompt() {
	params := url.Values{}

	params.Add("client_id", appId)
	params.Add("redirect_uri", fmt.Sprintf("http://localhost:%v/", listenPort))
	params.Add("response_type", "token")
	params.Add("scope", SCOPES)

	u, _ := url.Parse("https://www.donationalerts.com/oauth/authorize")
	u.RawQuery = params.Encode()

	os.WriteFile("Authorize Donationalerts.url", []byte(fmt.Sprintf("[InternetShortcut]\nURL=%v", u)), os.ModeAppend)

	internal.WarnLog.Println("")
	internal.WarnLog.Println("*** open 'Authorize Donationalerts' shortcut to connect the app")
	internal.WarnLog.Println("")
	internal.WarnLog.Printf("if it doesn't work, copy and paste this URL into your browser: %v", u)
}

func StartServer() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(fmt.Sprintf(":%v", listenPort), nil)
}

func GetUser() {
	req, _ := http.NewRequest("GET", USER_BASE_URL, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		internal.ErrLog.Printf("Error getting user info from donationalerts: %v", err)
		setState("connecting")
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code == 401 {
		accessToken = ""
		internal.Config.DaToken = ""
		internal.WriteConfig()
		Prompt()
		setState("connecting")
	} else if code != 200 {
		internal.ErrLog.Printf("Error getting user info from donationalerts: %v %v", code, http.StatusText(code))
		setState("connecting")
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			internal.ErrLog.Printf("Error reading user info from donationalerts: %v", err)
			setState("access_token_set")
		}
		userInfo := UserInfo{}
		err2 := json.Unmarshal(body, &userInfo)
		if err2 != nil {
			internal.ErrLog.Printf("Error reading user info from donationalerts: %v", err2)
			setState("access_token_set")
		}
		internal.InfoLog.Printf("authorized as %v", userInfo.Data.Name)

		userId = userInfo.Data.Id
		socketConnectionToken = userInfo.Data.SocketConnectionToken
		setState("user_set")
	}
}

func disconnectWs() {
	if conn != nil {
		conn.Close()
	}
}

func subscribe() {
	body := MakeSubscribeRequest(userId, clientKey)
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", SUBSCRIBE_URL, bytes.NewBuffer(b))
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		internal.ErrLog.Printf("Error subscribing to donationalerts channel: %v", err)
		setState("user_set")
		disconnectWs()
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code == 401 {
		accessToken = ""
		internal.Config.DaToken = ""
		internal.WriteConfig()
		Prompt()
		disconnectWs()
		setState("connecting")
	} else if code != 200 {
		internal.ErrLog.Printf("Error subscribing to donationalerts channel: %v %v", code, http.StatusText(code))
		disconnectWs()
		setState("user_set")
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			internal.ErrLog.Printf("Error reading subscribe response: %v", err)
			disconnectWs()
			setState("user_set")
		}
		subResp := SubscribeResponse{}
		err2 := json.Unmarshal(body, &subResp)
		if err2 != nil {
			internal.ErrLog.Printf("Error reading subscribe response: %v", err2)
			disconnectWs()
			setState("user_set")
		} else if len(subResp.Channels) < 1 {
			internal.ErrLog.Printf("No channels in subscribe response")
			disconnectWs()
			setState("user_set")
		}
		channel := subResp.Channels[0]
		internal.InfoLog.Printf("ready to connect to %v channel", channel.Channel)

		channelToken = channel.Token
		setState("subscribed")
	}
}

func connectWs() {
	d := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
	}
	c, _, err := d.Dial(WEBSOCKET_URL, nil)
	if err != nil {
		internal.ErrLog.Printf("error connecting to websocket: %v", err)
		reconnectAt = time.Now().Add(15 * time.Second)
		return
	}
	conn = c
	setState("ws_connected")
}

func sendInitMsg() {
	p := map[string]string{
		"token": socketConnectionToken,
	}
	msg := InitMsg{Id: nextId(), Params: p}
	err := conn.WriteJSON(msg)
	if err != nil {
		internal.ErrLog.Printf("error sending init message: %v", err)
		disconnectWs()
		setState("user_set")
	} else {
		setState("waiting_client")
	}
}

func sendConnectMsg() {
	p := map[string]string{
		"channel": fmt.Sprintf("$alerts:donation_%v", userId),
		"token":   channelToken,
	}
	msg := ConnectMsg{Id: nextId(), Method: 1, Params: p}
	err := conn.WriteJSON(msg)
	if err != nil {
		internal.ErrLog.Printf("error sending init message: %v", err)
		disconnectWs()
		setState("user_set")
	} else {
		setState("waiting_channel_connect")
	}
}

func readyToRead() bool {
	return state == "ws_connected" ||
		state == "waiting_client" ||
		state == "client_set" ||
		state == "subscribed" ||
		state == "waiting_channel_connect" ||
		state == "channel_connected"
}

func tryReadDonationAlert(r map[string]any) {
	data1, _ := r["data"]
	if data1 != nil {
		data2, _ := data1.(map[string]any)["data"]
		if data2 != nil {
			data, _ := data2.(map[string]any)
			if data != nil {
				name, _ := data["name"].(string)
				username, _ := data["username"].(string)
				amount, _ := data["amount"].(float64)
				currency, _ := data["currency"].(string)
				convertedAmount, _ := data["amount_in_user_currency"].(float64)
				internal.InfoLog.Printf("Received donation from %v/%v for %v %v (%v converted)",
					name, username, amount, currency, convertedAmount)
				if convertedAmount != 0.0 {
					vts.Control <- vts.ControlMsg{Msg: vts.SetParam, Value: multiplier * convertedAmount}
				}
			}
		}
	}
}

func read() {
	for {
		if state == "finished" {
			return
		} else if readyToRead() {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if state == "finished" {
					return
				}
				internal.ErrLog.Printf("read error: %v", err)
				conn.Close()
				setState("user_set")
			} else if msgType == websocket.TextMessage {
				strMsg := string(message)
				if internal.Config.LogDaMessages {
					internal.InfoLog.Printf("received text message: %v", strMsg)
				}
				response := ResponseMsg{}
				err := json.Unmarshal(message, &response)
				if err != nil {
					internal.ErrLog.Printf("can't parse reponse: %v", err)
				} else {
					result := response.Result
					client := result["client"]
					responseType := result["type"]

					if client != nil && client != "" {
						clientKey = client.(string)
						setState("client_set")
					} else if responseType != nil && responseType == 1.0 {
						internal.InfoLog.Printf("connected to donations channel")
						setState("channel_connected")
					} else {
						tryReadDonationAlert(result)
					}
				}
			} else {
				internal.InfoLog.Printf("received unknown message of type %v", msgType)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func Run(done chan<- int) {
	defer func() { done <- 1 }()
	go read()

	if accessToken == "" {
		Prompt()
	} else {
		setState("access_token_set")
	}
	go StartServer()

	for {
		select {
		case t := <-t.C:
			if state == "access_token_set" && t.After(reconnectAt) {
				GetUser()
			}
			if state == "user_set" && t.After(reconnectAt) {
				connectWs()
			}
			if state == "ws_connected" {
				sendInitMsg()
			}
			if state == "client_set" {
				subscribe()
			}
			if state == "subscribed" {
				sendConnectMsg()
			}
		case c := <-Control:
			if c == DONE {
				setState("finished")
				if conn != nil {
					conn.Close()
				}
				return
			}
		}
	}
}
