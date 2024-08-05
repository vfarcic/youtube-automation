package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Hugo struct{}

func (r *Hugo) Post(gist, title, date string) (string, error) {
	if gist == "N/A" {
		return "", nil
	}
	post := r.getPost(gist, title, date)
	return r.hugoFromMarkdown(gist, title, post)
}

func (r *Hugo) hugoFromMarkdown(filePath, title, post string) (string, error) {
	categoryDir := settings.Hugo.Path + "/" + strings.Replace(filepath.Dir(filePath), "manuscript", "content", 1)
	postDir := title
	postDir = strings.ReplaceAll(postDir, " ", "-")
	postDir = strings.ReplaceAll(postDir, "(", "")
	postDir = strings.ReplaceAll(postDir, ")", "")
	postDir = strings.ReplaceAll(postDir, ":", "")
	postDir = strings.ReplaceAll(postDir, "&", "")
	postDir = strings.ReplaceAll(postDir, "/", "-")
	postDir = strings.ReplaceAll(postDir, "'", "")
	postDir = strings.ReplaceAll(postDir, "!", "")
	postDir = strings.ToLower(postDir)
	fullDir := categoryDir + "/" + postDir
	os.Mkdir(fullDir, os.FileMode(0755))
	hugoPath := fullDir + "/_index.md"
	hugoPath = strings.Replace(hugoPath, "//", "/", -1)
	err := os.WriteFile(hugoPath, []byte(post), 0644)
	if err != nil {
		return "", err
	}
	return hugoPath, nil
}

func (r *Hugo) getPost(filePath, title, date string) string {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	content := fmt.Sprintf(`
+++
title = '%s'
date = %s:00+00:00
draft = false
+++

FIXME:

<!--more-->

{{< youtube FIXME: >}}

%s
`, title, date, string(contentBytes))
	return content
}
