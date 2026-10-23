package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Folder   string `json:"folder"`
	Contacts string `json:"contacts"`
}

var watchDir string

var contactsDir string

var lastFile string

func main() {

	cfg, err := loadConfig()
	if err != nil || len(cfg.Folder) == 0 {
		log.Fatalln("Ошибка загрузки config.json:", err)
		return
	}
	watchDir = cfg.Folder
	contactsDir = cfg.Contacts
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
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					if err := addDirRecursive(watcher, event.Name); err != nil {
						fmt.Println("Ошибка рекурсивного добавления директории:", err)
					}
				} else {
					processFile(event.Name)
				}
			}
		case err := <-watcher.Errors:
			fmt.Println("Ошибка watcher:", err)
		}
	}

}

func addDirRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if err := w.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
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

func showSystemNotificationTerminalNotifier(from, summary, file string) {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Ошибка получения пути к бинарнику:", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	// Формируем абсолютный путь для file, если он не абсолютный
	absFile := file
	if !strings.HasPrefix(file, "/") {
		absFile = filepath.Join(exeDir, file)
	}
	fileURL := url.URL{Scheme: "file", Path: absFile}
	args := []string{
		"-title", from,
		"-message", summary,
		"-open", fileURL.String(),
	}
	iconPath := filepath.Join(exeDir, "AppIcon.icns")
	if _, err := os.Stat(iconPath); err == nil {
		args = append(args, "-appIcon", iconPath)
	}
	if c := findContact(contactsDir, from); c != nil && c.Name != "" {
		args[1] = c.Name
		args = append(args, "-subtitle", from)
		if photo := c.photoFile(); photo != "" {
			photoURL := url.URL{Scheme: "file", Path: photo}
			args = append(args, "-contentImage", photoURL.String())
		}
	}
	err = exec.Command("terminal-notifier", args...).Run()
	if err != nil {
		fmt.Println("Ошибка terminal-notifier:", err)
	}
}
