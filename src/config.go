package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Testcenter struct {
		Street string `yaml:"street"`
		Plz    string `yaml:"plz"`
		City   string `yaml:"city"`
		Phone  string `yaml:"phone"`
		Email  string `yaml:"email"`
	} `yaml:"testcenter"`
	Ldnr struct {
		Prefix    string `yaml:"prefix"`
		Counter   int    `yaml:"counter"`
		NumLength int    `yaml:"numlength"`
	} `yaml:"ldnr"`
	Test struct {
		Hersteller string `yaml:"hersteller"`
		Pzn        string `yaml:"pzn"`
	} `yaml:"test"`
	Server struct {
		// Port is the local machine TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`

		// Host is the local machine IP Address to bind the HTTP Server to
		Host string `yaml:"host"`

		SavePath string `yaml:"save_path"`

		Timeout struct {
			// Server is the general server timeout to use
			// for graceful shutdowns
			Server time.Duration `yaml:"server"`

			// Write is the amount of time to wait until an HTTP server
			// write opperation is cancelled
			Write time.Duration `yaml:"write"`

			// Read is the amount of time to wait until an HTTP server
			// read operation is cancelled
			Read time.Duration `yaml:"read"`

			// Read is the amount of time to wait
			// until an IDLE HTTP session is closed
			Idle time.Duration `yaml:"idle"`
		} `yaml:"timeout"`
	} `yaml:"server"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func updateConfig(config Config) {
	log.Printf("Laufende Nummer erhöht zu: " + fmt.Sprintf("%x", config.Ldnr.Counter))
	d, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile("config.yml", d, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
