package notify

import (
	"fmt"
	"harmony/client/common"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/0xAX/notificator"
	"github.com/gen2brain/beeep"
)

const AppName = "Harmony"

func NotifyText(msg string) {
	notify(msg, []byte{}, common.TextType)
}

func NotifyImage(msg string, imgData []byte) {
	notify(msg, imgData, common.ImageType)
}

func notify(msg string, imgData []byte, dtype common.BufType) {
	tmpFile := ""
	if dtype == common.ImageType {
		tf, err := writeTempImage(imgData)
		if err != nil {
			log.Printf("Failed to write temp image: %v\n", err)
			return
		}
		go deleteTempFileAfterDelay(tf, 5*time.Second)
		tmpFile = tf
	}

	var err error
	switch runtime.GOOS {
	case "linux", "darwin":
		err = notifyUnix(msg, tmpFile)
	case "windows":
		err = notifyWindows(msg, tmpFile)
	default:
		log.Printf("Unsupported platform: %s\n", runtime.GOOS)
		return
	}

	if err != nil {
		log.Printf("Notification failed: %v\n", err)
	}
}

// Unix (beeep)
func notifyUnix(msg string, iconPath string) error {
	return beeep.Notify(AppName, msg, iconPath)
}

// Windows (go-toast)
func notifyWindows(msg string, iconPath string) error {
	notify := notificator.New(notificator.Options{
		DefaultIcon: "assets/icon.png",
		AppName:     "Harmony",
	})
	return notify.Push(AppName, msg, iconPath, notificator.UR_NORMAL)
}

func writeTempImage(data []byte) (string, error) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("harmony_icon_%d.png", rand.Intn(100)))
	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return "", err
	}
	if err := f.Sync(); err != nil {
		return "", err
	}

	return tmpFile, nil
}

func deleteTempFileAfterDelay(path string, delay time.Duration) {
	time.Sleep(delay)
	if err := os.Remove(path); err != nil {
		log.Printf("Failed to remove temp file: %s, error: %v\n", path, err)
	} else {
		log.Printf("Temp file deleted: %s\n", path)
	}
}
