package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Configuration struct {
	FrontendDir        string `json:"frontenddir"`
	PoolServiceBaseURL string `json:"poolservicebaseurl"`
}

func Setup() (Configuration, error) {
	var conf Configuration
	viper.SetConfigName("config")               // name of config file (without extension)
	viper.AddConfigPath("/etc/mtggameengine/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.mtggameengine") // call multiple times to add many search paths
	viper.AddConfigPath(".")                    // optionally look for config in the working directory
	err := viper.ReadInConfig()                 // Find and read the config file
	if err != nil {                             // Handle errors reading the config file
		return conf, fmt.Errorf("Fatal error config file: %s \n", err)
	}
	err = viper.Unmarshal(&conf)
	if err != nil { // Handle errors reading the config file
		return conf, fmt.Errorf("Fatal error config file: %s \n", err)
	}

	return conf, nil
}
