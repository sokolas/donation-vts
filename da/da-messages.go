package da

import "fmt"

type UserInfoData struct {
	Id                    int64  `json:"id"`
	Name                  string `json:"name"`
	SocketConnectionToken string `json:"socket_connection_token"`
}

type UserInfo struct {
	Data UserInfoData `json:"data"`
}

type InitMsg struct {
	Id     int64             `json:"id"`
	Params map[string]string `json:"params"`
}

type ResponseMsg struct {
	Id     int64          `json:"id"`
	Result map[string]any `json:"result"`
}

type SubscribeRequest struct {
	Client   string   `json:"client"`
	Channels []string `json:"channels"`
}

func MakeSubscribeRequest(userId int64, token string) SubscribeRequest {
	channels := []string{fmt.Sprintf("$alerts:donation_%v", userId)}
	return SubscribeRequest{Client: token, Channels: channels}
}

type ChannelAndToken struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

type SubscribeResponse struct {
	Channels []ChannelAndToken `json:"channels"`
}

type ConnectMsg struct {
	Id     int64             `json:"id"`
	Method int32             `json:"method"`
	Params map[string]string `json:"params"`
}
