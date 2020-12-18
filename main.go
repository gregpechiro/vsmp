package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

var windowW, windowH int32

func main() {
	sdl.Main(func() {
		if err := run(); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(2)
		}
	})
}

func run() error {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var err error

	sdl.Do(func() {
		err = sdl.Init(sdl.INIT_EVERYTHING)
	})
	if err != nil {
		return fmt.Errorf("could not initialize SDL: %v", err)
	}
	defer func() {
		sdl.Do(func() {
			sdl.Quit()
		})
	}()

	sdl.Do(func() {
		window, renderer, err = sdl.CreateWindowAndRenderer(800, 600, sdl.WINDOW_SHOWN /*|sdl.WINDOW_FULLSCREEN_DESKTOP*/)
	})
	if err != nil {
		return fmt.Errorf("could not create window: %v", err)
	}
	defer func() {
		sdl.Do(func() {
			window.Destroy()
		})
	}()

	defer func() {
		sdl.Do(func() {
			renderer.Destroy()
		})
	}()

	windowW, windowH = window.GetSize()

	scene, err := newScene(renderer)
	if err != nil {
		return fmt.Errorf("could not create scene: %v", err)
	}
	defer scene.destroy()

	events := make(chan sdl.Event)
	errc := scene.run(events)

	for {
		select {
		case events <- sdl.WaitEvent():
		case err := <-errc:
			return err
		}
	}
}
