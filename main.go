package main

import (
	"context"
	"flag"
	"fmt"
	"gcm/internal/appConfig"
	"gcm/internal/cloneCommand"
	"gcm/internal/cloneCommand/terminalView"
	. "gcm/internal/log"
	"gcm/internal/view"
	typex "gcm/type"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"time"

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
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	compositeView := view.NewCompositeView(make([]view.View, 0))
	startTime := time.Now()

	out := os.Stdout
	errorViewModel := view.NewErrorViewModel(GetLogFilePath())
	compositeView.AddFooter(view.NewErrorView(errorViewModel, out))

	clonedNowViewModel := terminalView.NewClonedNowViewModel()
	compositeView.AddFooter(terminalView.NewClonedNowView(clonedNowViewModel, out))

	compositeView.AddFooter(view.NewTimeElapsedView(startTime, out, time.Since))

	ctx, stopRenderLoop := context.WithCancel(context.Background())
	if isTTY {
		go view.StartTTYRenderLoop(compositeView, out, ctx, os.Stdout)
	}

	cloneView := terminalView.NewGitLabCloneView(out)
	compositeView.AddView(cloneView)

	cloneCommand.ExecuteCloneCommand(config, cloneView, clonedNowViewModel, errorViewModel.ErrorChannel)

	stopRenderLoop()

	if !isTTY {
		compositeView.Render(0)
	}

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
