package palette

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/BrianBland/go-hue"
)

const (
	CONFIGFILE = "palette.json"
	DEVICETYPE = "palette#Lark"
)

type Palette struct {
	*hue.User
}

type config struct {
	Username string `json:"username"`
}

func New(bridge *hue.Bridge) (*Palette, error) {
	user, err := bridge.CreateUser(DEVICETYPE, "")
	if err != nil {
		return nil, err
	}
	return &Palette{User: user}, nil
}

func LoadFromConfig(bridge *hue.Bridge) (*Palette, error) {
	var c config
	if configBytes, err := ioutil.ReadFile(CONFIGFILE); err != nil {
		return nil, err
	} else {
		err = json.Unmarshal(configBytes, &c)
		if err != nil {
			return nil, err
		}
		if isValid, err := bridge.IsValidUser(c.Username); err != nil {
			return nil, err
		} else if !isValid {
			return nil, errors.New("Invalid user")
		}
	}
	p := Palette{User: hue.NewUserWithBridge(c.Username, bridge)}
	return &p, nil
}

func (p *Palette) SaveToConfig() error {
	config := config{Username: p.Username}
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(CONFIGFILE, b, 0666)
}
