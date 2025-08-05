package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	gosxnotifier "github.com/deckarep/gosx-notifier"
	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
)

type Config struct {
	Folder string `json:"folder"`
}

var watchDir string

var lastFile string

func main() {

	cfg, err := loadConfig()
	if err != nil || len(cfg.Folder) == 0 {
		log.Fatalln("Ошибка загрузки config.json:", err)
		return
	}
	watchDir = cfg.Folder
	// showSystemNotification("Тест", "Это тестовое уведомление", "/Users/alien/Vault/Messages/Emails/invitations@linkedin.com/2025-08-05 19꞉21 I want to connect.eml")
	// showSystemNotificationTerminalNotifier("Тест Terminal Notifier", "Это уведомление через terminal-notifier", "/Users/alien/Vault/Messages/Emails/invitations@linkedin.com/2025-08-05 19꞉21 I want to connect.eml")
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Ошибка получения пути к бинарнику:", err)
		return
	}
	lastMod := getModTime(exePath)
	go func() {
		for {
			time.Sleep(2 * time.Second)
			if isUpdated(exePath, lastMod) {
				fmt.Println("Бинарник обновлён — перезапуск...")
				exec.Command(exePath).Start()
				os.Exit(0)
			}
		}
	}()
	watch()

}

func watch() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	err = filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Ошибка Walk:", err)
		return
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create != 0 {
				processFile(event.Name)
			}
		case err := <-watcher.Errors:
			fmt.Println("Ошибка watcher:", err)
		}
	}

}

func processFile(file string) {

	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		return
	}
	if file == lastFile {
		return
	}
	lastFile = file
	xtype := getXAttr(file, "type")
	if xtype != "notification" {
		return
	}
	from := getXAttr(file, "from")
	summary := getXAttr(file, "summary")
	if summary == "" || from == "" {
		return
	}
	showSystemNotificationTerminalNotifier(from, summary, file)
}

func getXAttr(file, attr string) string {

	out, err := exec.Command("xattr", "-p", attr, file).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func getModTime(path string) time.Time {

	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}

	return info.ModTime()

}

func isUpdated(path string, last time.Time) bool {

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.ModTime().After(last)

}

func loadConfig() (Config, error) {

	exePath, err := os.Executable()
	if err != nil {
		return Config{}, err
	}
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)

	return cfg, err
}

func showSystemNotification(from, summary, file string) {
	if isDarwin() {
		note := gosxnotifier.NewNotification(summary)
		note.Title = from
		note.Sound = gosxnotifier.Default
		note.Sender = "com.apple.stickies"
		note.Link = "file://" + file
		err := note.Push()
		if err != nil {
			fmt.Println("Ошибка gosx-notifier:", err)
		}
	} else {
		err := beeep.Notify(from, summary, "")
		if err != nil {
			fmt.Println("Ошибка beeep.Notify:", err)
		}
	}
}

func isDarwin() bool {
	return strings.Contains(strings.ToLower(runtime.GOOS), "darwin")
}

func showSystemNotificationTerminalNotifier(from, summary, file string) {
	// Формируем абсолютный путь для file, если он не абсолютный
	absFile := file
	if !strings.HasPrefix(file, "/") {
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			absFile = filepath.Join(exeDir, file)
		}
	}
	err := exec.Command(
		"terminal-notifier",
		"-title", from,
		"-message", summary,
		"-open", "file://"+absFile,
	).Run()
	if err != nil {
		fmt.Println("Ошибка terminal-notifier:", err)
	}
}
