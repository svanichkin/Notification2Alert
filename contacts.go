package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
)

type Contact struct {
	Name   string
	Emails []string
	Photo  []byte
	Ext    string
}

// findContact ищет в папке с vcf карточку, к которой привязана почта
func findContact(dir, from string) *Contact {

	if dir == "" {
		return nil
	}
	email := normalizeEmail(from)
	if email == "" {
		return nil
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.vcf"))
	if err != nil {
		return nil
	}
	for _, file := range files {
		for _, c := range parseVCF(file) {
			for _, e := range c.Emails {
				if e == email {
					return c
				}
			}
		}
	}

	return nil
}

func normalizeEmail(s string) string {

	s = strings.TrimSpace(s)
	if strings.Contains(s, "<") {
		if addr, err := mail.ParseAddress(s); err == nil {
			s = addr.Address
		}
	}

	return strings.ToLower(s)
}

func parseVCF(file string) []*Contact {

	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	var contacts []*Contact
	var cur *Contact
	var org, showAs string
	for _, line := range unfoldLines(f) {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		params := strings.Split(key, ";")
		name := strings.ToUpper(params[0])
		// item1.EMAIL -> EMAIL
		if i := strings.LastIndex(name, "."); i >= 0 {
			name = name[i+1:]
		}
		switch name {
		case "BEGIN":
			cur = &Contact{}
			org, showAs = "", ""
		case "END":
			if cur != nil {
				if cur.Name == "" || strings.EqualFold(showAs, "COMPANY") && org != "" {
					cur.Name = org
				}
				contacts = append(contacts, cur)
			}
			cur = nil
		}
		if cur == nil {
			continue
		}
		switch name {
		case "FN":
			cur.Name = unescape(value)
		case "ORG":
			org = unescape(strings.Split(value, ";")[0])
		case "X-ABSHOWAS":
			showAs = value
		case "EMAIL":
			cur.Emails = append(cur.Emails, strings.ToLower(strings.TrimSpace(value)))
		case "PHOTO":
			cur.Photo, cur.Ext = decodePhoto(params[1:], value)
		}
	}

	return contacts
}

// unfoldLines склеивает перенесённые строки (продолжение начинается с пробела или таба)
func unfoldLines(f *os.File) []string {

	var lines []string
	var cur strings.Builder
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			cur.WriteString(line[1:])
			continue
		}
		if cur.Len() > 0 {
			lines = append(lines, cur.String())
			cur.Reset()
		}
		cur.WriteString(line)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}

	return lines
}

func unescape(s string) string {

	r := strings.NewReplacer(`\,`, ",", `\;`, ";", `\n`, " ", `\N`, " ", `\\`, `\`)

	return strings.TrimSpace(r.Replace(s))
}

// decodePhoto поддерживает vCard 3.0 (ENCODING=b;TYPE=PNG) и 4.0 (data:image/png;base64,...)
func decodePhoto(params []string, value string) ([]byte, string) {

	ext := "png"
	if strings.HasPrefix(value, "data:") {
		meta, data, ok := strings.Cut(value, ",")
		if !ok || !strings.Contains(meta, ";base64") {
			return nil, ""
		}
		if strings.Contains(strings.ToLower(meta), "jpeg") || strings.Contains(strings.ToLower(meta), "jpg") {
			ext = "jpg"
		}
		value = data
	} else {
		for _, p := range params {
			p = strings.ToUpper(p)
			if strings.HasPrefix(p, "TYPE=") && strings.Contains(p, "JP") {
				ext = "jpg"
			}
		}
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, ""
	}

	return data, ext
}

// photoFile сохраняет аватарку в кеш и возвращает путь к файлу
func (c *Contact) photoFile() string {

	if len(c.Photo) == 0 {
		return ""
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(cacheDir, "Notification2Alert", "avatars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	sum := sha1.Sum(c.Photo)
	path := filepath.Join(dir, hex.EncodeToString(sum[:])+"."+c.Ext)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	if err := os.WriteFile(path, c.Photo, 0o644); err != nil {
		return ""
	}

	return path
}
