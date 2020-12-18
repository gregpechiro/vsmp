package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var errQuit = errors.New("quit")

type scene struct {
	frameTexture *sdl.Texture
	renderer     *sdl.Renderer
	quitMenu     *sdl.MessageBoxData
	config       *config
}

func newScene(renderer *sdl.Renderer) (*scene, error) {

	var frameTexture *sdl.Texture

	config, err := newConfig()
	if err != nil {
		return nil, err
	}

	scene := &scene{
		frameTexture: frameTexture,
		renderer:     renderer,
		quitMenu:     newQuitMenu(),
		config:       config,
	}

	if err := scene.paint(); err != nil {
		return nil, err
	}

	return scene, nil
}

func (scene *scene) run(events chan sdl.Event) <-chan error {
	errc := make(chan error)
	go func() {
		defer close(errc)

		ticker := time.NewTicker(scene.config.tick * time.Second)
		for {
			if scene.config.updateTick {
				ticker.Reset(scene.config.tick * time.Second)
				scene.config.updateTick = false
			}
			select {
			case event := <-events:
				if err := scene.handleEvent(event); err != nil {
					if err == errQuit {
						return
					}
					if err != nil {
						errc <- err
					}
				}
			case <-ticker.C:
				fmt.Printf("Frame Number: %d\n", scene.config.CurrentFrame)
				if err := scene.paint(); err != nil {
					errc <- err
				}
			}
		}
	}()

	return errc
}

func (scene *scene) handleEvent(event sdl.Event) error {
	switch typ := event.(type) {
	case *sdl.QuitEvent:
		return errQuit
	case *sdl.KeyboardEvent:
		if typ.State == 1 && typ.Keysym.Mod == 4160 {
			switch typ.Keysym.Sym {
			case 27:
				if err := scene.quit(); err != nil {
					return err
				}
			case 48, 49, 50, 51, 52, 53, 54, 55, 56, 57:
				tick := time.Duration(typ.Keysym.Sym - 48)
				if tick == 0 {
					tick = scene.config.DefaultTick
				}
				scene.config.tick = time.Duration(tick)
				scene.config.updateTick = true
			default:
			}
		}

	case *sdl.MouseMotionEvent, *sdl.WindowEvent, *sdl.CommonEvent, *sdl.MouseButtonEvent:
	default:
		log.Printf("unknown event %T: %v\n", event, &event)
	}
	return nil
}

func (scene *scene) quit() error {
	button, err := sdl.ShowMessageBox(scene.quitMenu)
	if err != nil {
		return err
	}
	if button == 1 {
		return errQuit
	}
	return nil
}

func (scene *scene) paint() error {
	var err error

	if err := scene.loadFrameTexture(); err != nil {
		return err
	}

	sdl.Do(func() {
		scene.renderer.Clear()
		err = scene.renderer.Copy(scene.frameTexture, nil, scene.getScaledRect())
	})
	if err != nil {
		return fmt.Errorf("could not copy frame: %v", err)
	}

	sdl.Do(func() {
		scene.renderer.Present()
	})

	scene.config.increaseFrame()

	scene.config.saveConfig()

	return nil
}

func (scene *scene) loadFrameTexture() error {
	var err error

	fileExists := true

	filePath := fmt.Sprintf("resources/imgs/%d.jpg", scene.config.CurrentFrame)
	if !scene.config.CacheImages {
		filePath = "resources/imgs/currentframe.jpg"
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	if fileExists && !scene.config.CacheImages {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("could not remove existing image: %v", err)
		}
		fileExists = false
	}

	if !fileExists {
		cmd := exec.Command("ffmpeg",
			"-i", scene.config.MovieFilePath,
			"-vf", fmt.Sprintf("select=gte(n\\, %d)", scene.config.CurrentFrame),
			"-vframes", "1",
			"-vsync", "0",
			filePath,
		)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not generate frame image: %v", err)
		}
	}

	sdl.Do(func() {
		scene.frameTexture.Destroy()
	})

	sdl.Do(func() {
		scene.frameTexture, err = img.LoadTexture(scene.renderer, filePath)
	})
	if err != nil {
		return fmt.Errorf("could not load frame image: %v", err)
	}

	return nil
}

func (scene *scene) getScaledRect() *sdl.Rect {
	limitW, limitH, _ := scene.renderer.GetOutputSize()
	_, _, textureW, textureH, _ := scene.frameTexture.Query()

	var rect *sdl.Rect
	if textureW > limitW || textureH > limitH || (textureW < limitW || textureH < limitH) {
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

		rect = &sdl.Rect{X: offsetX, Y: offsetY, W: dstW, H: dstH}
	}

	return rect
}

func (scene *scene) destroy() {
	sdl.Do(func() {
		scene.frameTexture.Destroy()
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
			{R: 62, G: 62, B: 62},
			/* [SDL_MESSAGEBOX_COLOR_TEXT] */
			{R: 221, G: 221, B: 221},
			/* [SDL_MESSAGEBOX_COLOR_BUTTON_BORDER] */
			{R: 20, G: 20, B: 20},
			/* [SDL_MESSAGEBOX_COLOR_BUTTON_BACKGROUND] */
			{R: 150, G: 150, B: 150},
			/* [SDL_MESSAGEBOX_COLOR_BUTTON_SELECTED] */
			{R: 65, G: 65, B: 65},
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
