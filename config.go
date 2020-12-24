package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type config struct {
	CurrentFrame  int
	MaxFrames     int
	tick          time.Duration
	DefaultTick   time.Duration
	updateTick    bool
	CacheImages   bool
	MovieFileName string
	movieFilePath string
	resourcesDir  string
}

func newConfig() (*config, error) {

	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not load base directory %v", err)
	}

	config := config{
		CurrentFrame: 0,
		MaxFrames:    0,
		DefaultTick:  120,
		updateTick:   false,
		CacheImages:  true,
		resourcesDir: filepath.Dir(executablePath) + "/resources",
	}

	b, err := ioutil.ReadFile(config.resourcesDir + "/config.json")
	if err != nil {
		return nil, fmt.Errorf("could not read config file %v", err)
	}

	if err := json.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file %v", err)
	}

	if config.MovieFileName == "" {
		return nil, fmt.Errorf("movie file name missing. Please set a movie file name in the config file")
	}

	config.movieFilePath = config.resourcesDir + "/movie/" + config.MovieFileName

	if _, err := os.Stat(config.movieFilePath); err != nil {
		return nil, fmt.Errorf("could not find movie %s. Please check the path in the config: %v", config.movieFilePath, err)
	}

	if config.MaxFrames < 1 {

		cmd := exec.Command("ffprobe",
			"-v", "error",
			"-select_streams", "v:0",
			"-show_entries", "stream=nb_frames",
			"-of", "default=nokey=1:noprint_wrappers=1",
			config.movieFilePath,
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(stdout)
		maxFramesStr := strings.TrimSpace(buf.String())

		if err := cmd.Wait(); err != nil {
			return nil, err
		}

		maxFrames, err := strconv.Atoi(maxFramesStr)
		if err != nil {
			return nil, err
		}

		config.MaxFrames = maxFrames
	}

	config.tick = config.DefaultTick

	return &config, nil
}

func (config *config) increaseFrame() {
	config.CurrentFrame++
	if config.CurrentFrame >= config.MaxFrames {
		config.CurrentFrame = 0
	}
}

func (config *config) saveConfig() {
	b, err := json.Marshal(config)
	if err != nil {
		log.Printf("Error --> json.Marshal(config): %v", err)
	}

	if err := ioutil.WriteFile(config.resourcesDir+"/config.json", b, 0644); err != nil {
		log.Printf("Error --> ioutil.WriteFile(\""+config.resourcesDir+"/config.json\", b, 0644): %v", err)
	}
}
