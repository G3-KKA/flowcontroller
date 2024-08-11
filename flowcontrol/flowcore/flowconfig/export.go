package flowcfg

import (
	"errors"
	"flowcontroller/config"

	_ "github.com/spf13/viper"
)

var c config.Config

const noexport_CONFIG_KEY = `gD/33YoHZP3BezxvWeGaIw==`

var (
	ErrIncorrectConfigKey = errors.New("incorrect config key")
)

// # Flow Controller Internal Config Initialiser
//
// # Do not try to use it, instead register your service and get config from metadata
//
// On incorrect key returns [ErrIncorrectConfigKey]
func ReadConfig(key string) (config.Config, error) {
	if key != noexport_CONFIG_KEY {
		return config.Config{}, ErrIncorrectConfigKey
	}
	err := initConfig()
	return c, err
}
