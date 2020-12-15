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
	var w *sdl.Window
	var r *sdl.Renderer
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
		w, r, err = sdl.CreateWindowAndRenderer(800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_FULLSCREEN_DESKTOP)
	})
	if err != nil {
		return fmt.Errorf("could not create window: %v", err)
	}
	defer func() {
		sdl.Do(func() {
			w.Destroy()
		})
	}()

	defer func() {
		sdl.Do(func() {
			r.Destroy()
		})
	}()

	windowW, windowH = w.GetSize()

	s, err := newScene(r)
	if err != nil {
		return fmt.Errorf("could not create scene: %v", err)
	}
	defer s.destroy()

	events := make(chan sdl.Event)
	errc := s.run(events)

	for {
		select {
		case events <- sdl.WaitEvent():
		case err := <-errc:
			return err
		}
	}
}
