package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/foomo/kraken"
)

const version = "0.4.1"

var flagVersion = flag.Bool("version", false, "display version")
var flagAddress = flag.String("address", "127.0.0.1:8888", "where to listen too like 127.0.0.1:8888")
var flagConfig = flag.String("config", "/etc/kraken/config.yaml", "where to find a config file")
var flagInsecure = flag.Bool("insecure", false, "disabled by default - controls whether a client verifies the server's certificate chain and host name")

// Config structure
type Config struct {
	Address   string                               `yaml:"address,omitempty"`
	Tentacles map[string]kraken.TentacleDefinition `yaml:"tentacles,omitempty"`
}

func readConfig(file string) (data Config, err error) {
	var config Config
	config.Address = *flagAddress

	if len(file) == 0 {
		return config, nil
	}
	dataBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return config, errors.New("could not read config file: " + err.Error())
	}

	if strings.HasSuffix(file, ".json") {
		err = json.Unmarshal(dataBytes, &config)
	} else if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
		err = yaml.Unmarshal(dataBytes, &config)
	} else {
		return config, errors.New("unsupported config file format, requires: .json, .yml or .yaml")
	}
	return config, err
}

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println("version", version)
		os.Exit(2)
	}

	fmt.Println("reading config file:", *flagConfig)
	config, err := readConfig(*flagConfig)
	if err != nil {
		log.Println("unable to read config file")
		os.Exit(1)
	}

	fmt.Print("I AM KRAKEN " + version + " at: " + config.Address + "\n")

	if *flagInsecure {
		fmt.Println("!!! not verifying peers !!!")
		http.DefaultClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	k := kraken.NewKraken()
	for name, tentacle := range config.Tentacles {
		k.GrowTentacle(name, tentacle.Bandwidth, tentacle.Retry)
	}
	kraken.NewServer(k).ListenAndServe(config.Address)
}
