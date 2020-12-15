package main

import (
	"time"
)

type config struct {
	current     int
	max         int
	tick        time.Duration
	updateTick  bool
	cacheImages bool
}

func newConfig() *config {
	return &config{
		current:     0,
		max:         5,
		tick:        5,
		updateTick:  false,
		cacheImages: true,
	}
}
