package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Repo struct{}

func (r *Repo) Update(repo, title, videoID string) error {
	cmdClone := exec.Command("gh", "repo", "clone", repo)
	_, err := cmdClone.CombinedOutput()
	if err != nil {
		return err
	}
	readmePath := fmt.Sprintf("%s/README.md", repo)
	file, err := os.OpenFile(readmePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		file.Close()
		return err
	}
	defer file.Close()
	contentTitle := "# Demo Manifests and Code Used in DevOps Toolkit Videos"
	contentVideo := fmt.Sprintf("[![%s](https://img.youtube.com/vi/%s/0.jpg)](https://youtu.be/%s)", title, videoID, videoID)
	content := fmt.Sprintf("%s\n\n%s", contentTitle, contentVideo)
	_, err = io.WriteString(file, content)
	if err != nil {
		return err
	}
	cmdGit := []*exec.Cmd{
		exec.Command("git", "add", "."),
		exec.Command("git", "commit", "-m", "Update README"),
		exec.Command("git", "push"),
	}
	for _, cmd := range cmdGit {
		cmd.Dir = repo
		_, err = cmd.CombinedOutput()
		if err != nil {
			return err
		}
	}
	cmdDelete := exec.Command("rm", "-rf", repo)
	_, err = cmdDelete.CombinedOutput()
	if err != nil {
		return err
	}
	return err
}

func (r *Repo) GetAnimations(filePath string) (animations, sections []string, err error) {
	if strings.HasSuffix(filePath, ".sh") {
		return r.getAnimationsFromScript(filePath)
	}
	return r.getAnimationsFromMarkdown(filePath)
}

// TODO: Remove
func (r *Repo) getAnimationsFromScript(filePath string) (animations, sections []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, " ", " ")
		if strings.HasPrefix(line, "#") && strings.HasSuffix(line, "#") && !strings.HasPrefix(line, "##") {
			foundIt := false
			for _, value := range []string{"# [[title]] #", "# Intro #", "# Setup #", "# Destroy #"} {
				if line == value {
					foundIt = true
					break
				}
			}
			if !foundIt {
				line = strings.ReplaceAll(line, "#", "")
				line = strings.TrimSpace(line)
				line = fmt.Sprintf("Section: %s", line)
				animations = append(animations, line)
				sections = append(sections, line)
			}
		} else if strings.HasPrefix(line, "# TODO:") {
			line = strings.ReplaceAll(line, "# TODO:", "")
			line = strings.TrimSpace(line)
			animations = append(animations, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return animations, sections, nil
}

func (r *Repo) getAnimationsFromMarkdown(filePath string) (animations, sections []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, " ", " ")
		if strings.HasPrefix(line, "TODO:") {
			line = strings.ReplaceAll(line, "TODO:", "")
			line = strings.TrimSpace(line)
			animations = append(animations, line)
		} else if strings.HasPrefix(line, "## ") {
			containsAny := false
			for _, value := range []string{"## Intro", "## Setup", "## Destroy"} {
				if line == value {
					containsAny = true
					break
				}
			}
			if !containsAny {
				line = strings.Replace(line, "## ", "", 1)
				line = strings.TrimSpace(line)
				line = fmt.Sprintf("Section: %s", line)
				animations = append(animations, line)
				sections = append(sections, line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return animations, sections, nil
}

func (r *Repo) CleanupGist(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines, outputLines []string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, " ", " ")
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for index := 0; index < len(lines); index++ {
		line := lines[index]
		if strings.HasPrefix(line, "##") { // Headers
			for headerIndex := 0; headerIndex <= 2; headerIndex++ {
				line = lines[index]
				outputLines = append(outputLines, line)
				index++
			}
			outputLines = append(outputLines, "")
		} else if strings.HasPrefix(line, "# FIXME:") { // Comments
			outputLines = append(outputLines, strings.Replace(line, "# FIXME:", "#", 1))
			outputLines = append(outputLines, "")
		} else if !strings.HasPrefix(line, "#") && len(line) > 0 { // Code
			outputLines = append(outputLines, line)
			if !strings.HasSuffix(line, "\\") {
				outputLines = append(outputLines, "")
			}
		} else if index <= 4 { // Title & Additional Info
			outputLines = append(outputLines, line)
		}
	}
	// Remove empty sections
	lines = outputLines
	outputLines = []string{}
	for index := 0; index < len(lines); index++ {
		line := lines[index]
		if len(lines) >= index+5 && strings.HasPrefix(line, "##") && strings.HasPrefix(lines[index+2], "##") && !strings.HasPrefix(lines[index+4], "##") {
			outputLines = append(outputLines, lines[index:index+3]...)
			index += 2
		} else if !strings.HasSuffix(line, "#") || line == "# [[title]] #" || line == "[[title]]" {
			outputLines = append(outputLines, line)
		}
	}

	err = os.WriteFile(filePath, []byte(strings.Join(outputLines, "\n")), 0644)
	return err
}
