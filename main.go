package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"tools/internal/appConfig"
	"tools/internal/cloneCommand"
	. "tools/internal/log"
	typex "tools/type"

	"gopkg.in/yaml.v2"
)

func main() {

	//f, _ := os.Create("trace.out")
	//defer f.Close()
	//trace.Start(f)
	//defer trace.Stop()

	// Process parameters
	var verbose = typex.NullableBool{}
	flag.Var(&verbose, "verbose", "Print verbose output")
	flag.Parse()
	InitLogger(verbose.Val(false))

	config, err := loadConfig("workingCopies.yaml")

	if err != nil {
		Log.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	cloneCommand.ExecuteCloneCommand(config)
}

func loadConfig(configFileName string) (*appConfig.AppConfig, error) {
	configFilePath := filepath.Join("./", configFileName)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %v", err)
		}
		configFilePath = filepath.Join(homeDir, configFileName)
		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found in current directory or home directory")
		}
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var config appConfig.AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %v", err)
	}

	return &config, nil
}
