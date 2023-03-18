package internal

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
)

type AppConfig struct {
	LogVtsMessages   bool    `json:"logVtsMessages"`
	LogDaMessages    bool    `json:"logDaMessages"`
	VtsAddr          string  `json:"vtsAddr"`
	VtsToken         string  `json:"vtsToken"`
	CustomParam      string  `json:"customParam"`
	ParamDescription string  `json:"paramDescription"`
	StayTime         int64   `json:"stayTime"`
	DecayTime        int64   `json:"decayTime"`
	Multiplier       float64 `json:"multiplier"`
	AddParam         bool    `json:"addParam"`
	DaToken          string  `json:"daToken"`
	DaPort           int32   `json:"daPort"`
	DaAppId          string  `json:"daAppId"`
	AutoOpenUi       bool    `json:"autoOpenUi"`
}

type AppConfigDto struct {
	LogVtsMessages   *bool    `json:"logVtsMessages"`
	LogDaMessages    *bool    `json:"logDaMessages"`
	VtsAddr          *string  `json:"vtsAddr"`
	VtsToken         *string  `json:"vtsToken"`
	CustomParam      *string  `json:"customParam"`
	ParamDescription *string  `json:"paramDescription"`
	StayTime         *int64   `json:"stayTime"`
	DecayTime        *int64   `json:"decayTime"`
	Multiplier       *float64 `json:"multiplier"`
	AddParam         *bool    `json:"addParam"`
	DaToken          *string  `json:"daToken"`
	DaPort           *int32   `json:"daPort"`
	DaAppId          *string  `json:"daAppId"`
	AutoOpenUi       *bool    `json:"autoOpenUi"`
}

var Config AppConfig

func CreateDto() AppConfigDto {
	t := Config
	r := AppConfigDto{
		LogVtsMessages:   &t.LogVtsMessages,
		LogDaMessages:    &t.LogDaMessages,
		VtsAddr:          &t.VtsAddr,
		VtsToken:         &t.VtsToken,
		CustomParam:      &t.CustomParam,
		ParamDescription: &t.ParamDescription,
		StayTime:         &t.StayTime,
		DecayTime:        &t.DecayTime,
		Multiplier:       &t.Multiplier,
		AddParam:         &t.AddParam,
		DaToken:          &t.DaToken,
		AutoOpenUi:       &t.AutoOpenUi,
		/*&t.DaPort,
		&t.DaAppId,*/
	}
	return r
}

func MergeDto(c AppConfigDto) AppConfig {
	r := Config
	if c.LogVtsMessages != nil {
		r.LogVtsMessages = *c.LogVtsMessages
	}
	if c.LogDaMessages != nil {
		r.LogDaMessages = *c.LogDaMessages
	}
	if c.VtsAddr != nil {
		r.VtsAddr = *c.VtsAddr
	}
	if c.VtsToken != nil {
		r.VtsToken = *c.VtsToken
	}
	if c.CustomParam != nil {
		r.CustomParam = *c.CustomParam
	}
	if c.ParamDescription != nil {
		r.ParamDescription = *c.ParamDescription
	}
	if c.StayTime != nil {
		r.StayTime = *c.StayTime
	}
	if c.DecayTime != nil {
		r.DecayTime = *c.DecayTime
	}
	if c.Multiplier != nil {
		r.Multiplier = *c.Multiplier
	}
	if c.AddParam != nil {
		r.AddParam = *c.AddParam
	}
	if c.DaToken != nil {
		r.DaToken = *c.DaToken
	}
	if c.DaPort != nil {
		r.DaPort = *c.DaPort
	}
	if c.DaAppId != nil {
		r.DaAppId = *c.DaAppId
	}
	if c.AutoOpenUi != nil {
		r.AutoOpenUi = *c.AutoOpenUi
	}
	return r
}

func DumpConfig() {
	s := fmt.Sprintf("%#v", Config)
	if Config.VtsToken != "" {
		s = strings.Replace(s, Config.VtsToken, fmt.Sprintf("***(%v)", len(Config.VtsToken)), 1)
	}
	if Config.DaToken != "" {
		s = strings.Replace(s, Config.DaToken, fmt.Sprintf("***(%v)", len(Config.DaToken)), 1)
	}
	InfoLog.Println(s)
}

func ReadConfig() {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		WarnLog.Println("Error reading config.json: " + err.Error())
	} else {
		err2 := json.Unmarshal(bytes, &Config)
		if err2 != nil {
			WarnLog.Println("Error parsing config.json: " + err2.Error())
		}
	}
}

func WriteConfig() {
	bytes, err := json.MarshalIndent(Config, "", "\t")
	if err != nil {
		WarnLog.Printf("Error writing config.json: %v", err)
		return
	}
	err = os.WriteFile("config.json", bytes, fs.ModeAppend)
	if err != nil {
		WarnLog.Printf("Error writing config.json: %v", err)
	}
}
