//go:build linux

package hotkey

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
)

// Register registers a keyboard shortcut and invokes onPress when the combo is pressed.
func Register(combo string, onPress func()) {
	if onPress == nil {
		return
	}
	if combo == "" {
		combo = "Ctrl+V"
	}
	combo = strings.ToLower(strings.ReplaceAll(combo, " ", ""))

	if err := keyboard.Open(); err != nil {
		panic(fmt.Sprintf("open keyboard: %v", err))
	}
	defer keyboard.Close()

	keys, err := keyboard.GetKeys(10)
	if err != nil {
		panic(fmt.Sprintf("get keys: %v", err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	for {
		select {
		case <-stop:
			return
		case ev := <-keys:
			if ev.Err != nil {
				continue
			}

			matched := false
			switch combo {
			case "ctrl+v":
				matched = ev.Key == keyboard.KeyCtrlV || ev.Rune == 'v'
			case "ctrl+alt+v":
				matched = ev.Key == keyboard.KeyCtrlV || ev.Rune == 'v'
			case "v":
				matched = ev.Rune == 'v'
			default:
				matched = ev.Rune == 'v'
			}

			if matched {
				onPress()
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
