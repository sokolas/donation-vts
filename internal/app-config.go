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
}

var Config AppConfig

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
