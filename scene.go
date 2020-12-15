package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var quitError = errors.New("quit")

type scene struct {
	bg       *sdl.Texture
	r        *sdl.Renderer
	quitMenu *sdl.MessageBoxData
	config   *config
}

func newScene(r *sdl.Renderer) (*scene, error) {

	var bg *sdl.Texture
	var err error

	sdl.Do(func() {
		file := fmt.Sprintf("resources/imgs/0%d.jpg", 0)
		bg, err = img.LoadTexture(r, file)
	})
	if err != nil {
		return nil, fmt.Errorf("could not load image: %v", err)
	}

	s := &scene{
		bg:       bg,
		r:        r,
		quitMenu: newQuitMenu(),
		config:   newConfig(),
	}

	if err := s.paint(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *scene) run(events chan sdl.Event) <-chan error {
	errc := make(chan error)
	go func() {
		defer close(errc)

		ticker := time.NewTicker(s.config.tick * time.Second)
		for {
			if s.config.updateTick {
				ticker.Reset(s.config.tick * time.Second)
				s.config.updateTick = false
			}
			select {
			case e := <-events:
				if err := s.handleEvent(e); err != nil {
					if err == quitError {
						return
					}
					if err != nil {
						errc <- err
					}
				}
			case <-ticker.C:
				fmt.Println("tick")
				if err := s.paint(); err != nil {
					errc <- err
				}
			}
		}
	}()

	return errc
}

func (s *scene) handleEvent(event sdl.Event) error {
	switch t := event.(type) {
	case *sdl.QuitEvent:
		return quitError
	case *sdl.KeyboardEvent:
		if t.State == 1 && t.Keysym.Mod == 4160 {
			switch t.Keysym.Sym {
			case 27:
				if err := s.quit(); err != nil {
					return err
				}
			case 48, 49, 50, 51, 52, 53, 54, 55, 56, 57:
				tick := t.Keysym.Sym - 48
				if tick == 0 {
					tick = 10
				}
				s.config.tick = time.Duration(tick)
				s.config.updateTick = true
			default:
			}
		}

	case *sdl.MouseMotionEvent, *sdl.WindowEvent, *sdl.CommonEvent, *sdl.MouseButtonEvent:
	default:
		log.Printf("unknown event %T: %v\n", event, &event)
	}
	return nil
}

func (s *scene) quit() error {
	button, err := sdl.ShowMessageBox(s.quitMenu)
	if err != nil {
		return err
	}
	if button == 1 {
		return quitError
	}
	return nil
}

func (s *scene) paint() error {
	var err error

	sdl.Do(func() {
		s.bg.Destroy()
	})

	sdl.Do(func() {
		file := fmt.Sprintf("resources/imgs/0%d.jpg", s.config.current)
		s.bg, err = img.LoadTexture(s.r, file)
	})
	if err != nil {
		return fmt.Errorf("could not load image: %v", err)
	}

	limitW, limitH, _ := s.r.GetOutputSize()
	_, _, textureW, textureH, _ := s.bg.Query()

	var dstRect *sdl.Rect
	if textureW > limitW || textureH > limitH {
		scale := float64(limitW) / float64(textureW)
		scaleH := float64(limitH) / float64(textureH)

		if scaleH < scale {
			scale = scaleH
		}

		dstW := int32(float64(textureW) * scale)
		dstH := int32(float64(textureH) * scale)

		var offsetX int32
		var offsetY int32

		if limitW > dstW {
			offsetX = (limitW - dstW) / 2
		}

		if limitH > dstH {
			offsetY = (limitH - dstH) / 2
		}

		dstRect = &sdl.Rect{offsetX, offsetY, dstW, dstH}
	}

	sdl.Do(func() {
		s.r.Clear()
		err = s.r.Copy(s.bg, nil, dstRect)
	})
	if err != nil {
		return fmt.Errorf("could not copy background: %v", err)
	}

	sdl.Do(func() {
		s.r.Present()
	})

	if s.config.current >= s.config.max {
		s.config.current = 0
	} else {
		s.config.current++
	}

	return nil
}

func (s *scene) destroy() {
	sdl.Do(func() {
		s.bg.Destroy()
	})
}

func newQuitMenu() *sdl.MessageBoxData {
	buttons := []sdl.MessageBoxButtonData{
		{
			Flags:    sdl.MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT,
			ButtonID: 0,
			Text:     "No",
		},
		{
			Flags:    sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT,
			ButtonID: 1,
			Text:     "Yes",
		},
	}
	color := &sdl.MessageBoxColorScheme{
		Colors: [5]sdl.MessageBoxColor{
			/* .colors (.r, .g, .b) */
			/* [SDL_MESSAGEBOX_COLOR_BACKGROUND] */
			{R: 255, G: 0, B: 0},
			/* [SDL_MESSAGEBOX_COLOR_TEXT] */
			{R: 0, G: 255, B: 0},
			/* [SDL_MESSAGEBOX_COLOR_BUTTON_BORDER] */
			{R: 255, G: 255, B: 0},
			/* [SDL_MESSAGEBOX_COLOR_BUTTON_BACKGROUND] */
			{R: 0, G: 0, B: 255},
			/* [SDL_MESSAGEBOX_COLOR_BUTTON_SELECTED] */
			{R: 255, G: 0, B: 255},
		},
	}

	return &sdl.MessageBoxData{
		Flags:       sdl.MESSAGEBOX_INFORMATION,
		Window:      nil,
		Title:       "Quit",
		Message:     "Are you sure you would like to quit?",
		Buttons:     buttons,
		ColorScheme: color,
	}
}
