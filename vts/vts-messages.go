package vts

import "github.com/google/uuid"

const (
	API_NAME         = "VTubeStudioPublicAPI"
	API_VERSION      = "1.0"
	PLUGIN_NAME      = "donationalerts-param-plugin"
	PLUGIN_DEVELOPER = "sokolas"
)

type VTSMesssage struct {
	ApiName     string `json:"apiName"`
	ApiVersion  string `json:"apiVersion"`
	RequestId   string `json:"requestId"`
	MessageType string `json:"messageType"`
	Data        any    `json:"data"`
}

type TokenRequestData struct {
	PluginName      string `json:"pluginName"`
	PluginDeveloper string `json:"pluginDeveloper"`
}

type TokenAuthData struct {
	PluginName      string `json:"pluginName"`
	PluginDeveloper string `json:"pluginDeveloper"`
	Token           string `json:"authenticationToken"`
}

type CreateCustomParamData struct {
	ParamName    string  `json:"parameterName"`
	Explanation  string  `json:"explanation"`
	Min          int32   `json:"min"`
	Max          int32   `json:"max"`
	DefaultValue float64 `json:"defaultValue"`
}

type InjectCustomParamData struct {
	FaceFound bool         `json:"faceFound"`
	Mode      string       `json:"mode"`
	Values    []ParamValue `json:"parameterValues"`
}

type ParamValue struct {
	Id    string  `json:"id"`
	Value float64 `json:"value"`
}

func TokenRequest() VTSMesssage {
	return VTSMesssage{
		ApiName:     API_NAME,
		ApiVersion:  API_VERSION,
		RequestId:   uuid.NewString(),
		MessageType: "AuthenticationTokenRequest",
		Data: TokenRequestData{
			PluginName:      PLUGIN_NAME,
			PluginDeveloper: PLUGIN_DEVELOPER,
		},
	}
}

func AuthRequest(token string) VTSMesssage {
	return VTSMesssage{
		ApiName:     API_NAME,
		ApiVersion:  API_VERSION,
		RequestId:   uuid.NewString(),
		MessageType: "AuthenticationRequest",
		Data: TokenAuthData{
			PluginName:      PLUGIN_NAME,
			PluginDeveloper: PLUGIN_DEVELOPER,
			Token:           token,
		},
	}
}

func CreateParamRequest(name string, description string) VTSMesssage {
	return VTSMesssage{
		ApiName:     API_NAME,
		ApiVersion:  API_VERSION,
		RequestId:   uuid.NewString(),
		MessageType: "ParameterCreationRequest",
		Data: CreateCustomParamData{
			ParamName:    name,
			Explanation:  description,
			Min:          0,
			Max:          100,
			DefaultValue: 0,
		},
	}
}

func SetParamRequest(name string, value float64) VTSMesssage {
	values := make([]ParamValue, 1)
	values[0] = ParamValue{name, value}

	return VTSMesssage{
		ApiName:     API_NAME,
		ApiVersion:  API_VERSION,
		RequestId:   uuid.NewString(),
		MessageType: "InjectParameterDataRequest",
		Data: InjectCustomParamData{
			FaceFound: false,
			Mode:      "set",
			Values:    values,
		},
	}
}
