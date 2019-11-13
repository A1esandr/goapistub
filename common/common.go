package common

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	modeConfigUnknown = iota
	modeConfigFile
	modeConfigEnv
)

type (
	//HTTPConfig part of configuration
	HTTPConfig struct {
		Timeout              int `json:"timeout"`
		RequestClientTimeout int `json:"request_timeout"`
	}

	//Config - application configuration
	Config struct {
		mode           int
		fileNameConfig string

		Listen string     `json:"listen"`
		HTTP   HTTPConfig `json:"http"`
	}
)

var (
	lookupEnv = os.LookupEnv
	//ReaderFile define for test
	ReaderFile = ioutil.ReadFile

	//LoadConfig - returns Config. For the environment variable CONFIG_ENV="1"
	//from  environments else from config.json
	LoadConfig = loadConfig
)

//NewConfig returns new config
func NewConfig() *Config {
	return &Config{
		mode: modeConfigUnknown,
	}
}

//Check configuration
func (c *Config) Check() error {
	// Checks
	return nil
}

//LoadConfig - returns Config. For the environment variable CONFIG_ENV="1"
//from  environments else from config.json
func loadConfig() (*Config, error) {
	config := &Config{
		mode: modeConfigUnknown,
	}

	if val, ok := lookupEnv("CONFIG_ENV"); ok && val == "1" {
		config.mode = modeConfigEnv
		if err := config.loadFromEnv(); err != nil {
			return nil, err
		}
	} else {

		if err := config.loadFromFile(); err != nil {
			return nil, err
		}
		config.mode = modeConfigFile
	}
	config.setDefaultValues()
	return config, nil
}

func (c *Config) setDefaultValues() {
	if c.HTTP.Timeout < 0 {
		c.HTTP.Timeout = 0
	}
}

func (c *Config) loadFromFile() error {
	fileName := "config/config.json"
	if flag.NArg() != 0 {
		fileName = flag.Arg(0)
	}
	data, err := ReaderFile(fileName)
	if err != nil {
		return err
	}
	c.fileNameConfig = fileName
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("Error to parse config %s", err)
	}
	return nil
}

func envGet(key string) string {
	env, ok := lookupEnv(key)
	if !ok {
		return ""
	}
	return env
}
func envGetInt(key string) (int, error) {
	env, ok := lookupEnv(key)
	if !ok {
		return 0, nil
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		return 0, fmt.Errorf("could not read env %s: %s", key, err)
	}
	return val, nil
}

func (c *Config) loadFromEnv() error {
	var err error

	c.Listen = envGet("LISTEN")
	c.HTTP.Timeout, err = envGetInt("HTTP_CLIENT_TIMEOUT")
	if err != nil {
		return err
	}
	c.HTTP.RequestClientTimeout, err = envGetInt("HTTP_REQUEST_TIMEOUT")
	if err != nil {
		return err
	}

	return nil
}

//Informer provides configuration information
type Informer interface {
	Info(indent string) string
}

//Info returns configuration information
func (c *Config) Info() string {
	var w strings.Builder
	informers := []Informer{
		c.HTTP,
	}
	indent := "\t"
	modes := map[int]string{modeConfigUnknown: "unknown", modeConfigFile: "file", modeConfigEnv: "env"}
	w.WriteString("Environment:\n" + indent + "mode:" + modes[c.mode] + "\n") //nolint
	if c.mode == modeConfigFile {
		w.WriteString(indent + "file:" + c.fileNameConfig + "\n") //nolint
	}

	for _, informer := range informers {
		if _, err := w.WriteString(informer.Info(indent)); err != nil {
			panic(err)
		}
	}
	return w.String()
}

//Info returns configuration information
func (h HTTPConfig) Info(indent string) string {
	return fmt.Sprintf("HTTP client:\n%stimeout:%d\n", indent, h.Timeout)
}
