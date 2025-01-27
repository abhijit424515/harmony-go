package notify

import "github.com/gen2brain/beeep"

const AppName = "Harmony"

func Notify(msg string) {
	beeep.Notify(AppName, msg, "assets/icon.png")
}
