package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
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
	MovieFilePath string
}

func newConfig() (*config, error) {

	config := config{
		CurrentFrame: 0,
		MaxFrames:    0,
		DefaultTick:  120,
		updateTick:   false,
		CacheImages:  true,
	}

	b, err := ioutil.ReadFile("resources/config.json")
	if err != nil {
		return nil, fmt.Errorf("could not read config file %v", err)
	}

	if err := json.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file %v", err)
	}

	if config.MovieFilePath == "" {
		return nil, fmt.Errorf("movie path missing. Please set a movie path om the config file")
	}

	if config.MaxFrames < 1 {

		cmd := exec.Command("ffprobe",
			"-v", "error",
			"-select_streams", "v:0",
			"-show_entries", "stream=nb_frames",
			"-of", "default=nokey=1:noprint_wrappers=1",
			config.MovieFilePath,
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

	fmt.Printf("Loaded Config: %v\n", config)

	return &config, nil
}

func (config *config) increaseFrame() {
	if config.CurrentFrame >= config.MaxFrames {
		config.CurrentFrame = 0
	} else {
		config.CurrentFrame++
	}
}

func (config *config) saveConfig() {
	b, err := json.Marshal(config)
	if err != nil {
		log.Printf("Error --> json.Marshal(config): %v", err)
	}

	if err := ioutil.WriteFile("resources/config.json", b, 0644); err != nil {
		log.Printf("Error --> ioutil.WriteFile(\"resources/config.json\", b, 0644): %v", err)
	}
}
