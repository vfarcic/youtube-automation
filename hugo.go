package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Hugo struct{}

func (r *Hugo) Post(gist, title, projectName, projectUrl, relatedVideos string) (string, error) {
	if gist == "N/A" {
		return "", nil
	}
	return r.hugoFromMarkdown(gist, title, projectName, projectUrl, relatedVideos)
}

func (r *Hugo) hugoFromMarkdown(filePath, title, projectName, projectUrl, relatedVideos string) (string, error) {
	titleBorder := strings.Repeat("#", len(title)+4)
	gist := fmt.Sprintf("%s\n# %s #\n%s\n", titleBorder, title, titleBorder)
	additionalInfo := "# Additional Info:\n"
	if len(projectUrl) > 0 {
		additionalInfo = fmt.Sprintf("%s# - %s: %s\n", additionalInfo, projectName, projectUrl)
	}
	if len(relatedVideos) > 0 {
		a := strings.Split(relatedVideos, "\n")
		for _, t := range a {
			additionalInfo = fmt.Sprintf("%s# - %s\n", additionalInfo, t)
		}
	}
	if len(additionalInfo) > 0 {
		gist = fmt.Sprintf("%s\n%s\n", gist, strings.TrimRight(additionalInfo, "\n"))
	}
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	sections := make(map[string][]string)
	sh := false
	section := ""
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, " ", " ")
		if strings.HasPrefix(line, "## ") {
			section = strings.Replace(line, "## ", "", 1)
		} else if strings.HasPrefix(line, "FIXME:") {
			sections[section] = append(sections[section], strings.Replace(line, "FIXME:", "#", 1))
		} else if line == "```sh" || line == "```bash" {
			sh = true
		} else if line == "```" {
			sh = false
		} else if sh && len(line) > 0 {
			sections[section] = append(sections[section], line)
		}
	}
	for section, lines := range sections {
		if len(lines) > 0 {
			decoration := strings.Repeat("#", len(section)+4)
			gist = fmt.Sprintf("%s\n%s\n# %s #\n%s\n", gist, decoration, section, decoration)
			for _, line := range lines {
				gist = fmt.Sprintf("%s\n%s", gist, line)
				if !strings.HasSuffix(line, "\\") {
					gist = fmt.Sprintf("%s\n", gist)
				}
			}
		}
	}
	categoryDir := settings.Hugo.Path + "/" + strings.Replace(filepath.Dir(filePath), "manuscript", "content", 1)
	postDir := title
	postDir = strings.ReplaceAll(postDir, " ", "-")
	postDir = strings.ReplaceAll(postDir, "(", "")
	postDir = strings.ReplaceAll(postDir, ")", "")
	postDir = strings.ReplaceAll(postDir, ":", "")
	postDir = strings.ToLower(postDir)
	fullDir := categoryDir + "/" + postDir
	os.Mkdir(fullDir, os.FileMode(0755))
	hugoPath := fullDir + "/_index.md"
	hugoPath = strings.Replace(hugoPath, "//", "/", -1)
	// err = os.WriteFile(hugoPath, []byte(gist), 0644)
	// if err != nil {
	// 	return "", err
	// }
	println(hugoPath)
	return hugoPath, nil
}
