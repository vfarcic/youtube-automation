package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Repo struct{}

func (r *Repo) Gist(gist, title, projectName, projectUrl, relatedVideos string) error {
	if len(title) == 0 {
		errorMessage = "Please set the title of the video first"
		return fmt.Errorf(errorMessage)
	}
	data, err := os.ReadFile(gist)
	if err != nil {
		return err
	}
	titleBorder := strings.Repeat("#", len(title)+4)
	newTitle := fmt.Sprintf("%s\n# %s #\n%s", titleBorder, title, titleBorder)
	additionalInfo := ""
	if len(projectUrl) > 0 {
		additionalInfo = fmt.Sprintf("# - %s: %s\n", projectName, projectUrl)
	}
	if len(relatedVideos) > 0 {
		a := strings.Split(relatedVideos, "\n")
		for _, t := range a {
			additionalInfo = fmt.Sprintf("%s# - %s\n", additionalInfo, t)
		}
	}
	if len(additionalInfo) > 0 {
		additionalInfo = strings.TrimRight(additionalInfo, "\n")
	}
	modifiedData := strings.Replace(string(data), "# [[title]] #", newTitle, -1)
	modifiedData = strings.Replace(modifiedData, "# - [[additional-info]]", additionalInfo, -1)
	err = os.WriteFile(gist, []byte(modifiedData), 0644)
	if err != nil {
		return err
	}
	cmd := exec.Command("gh", "gist", "create", "--public", gist)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	gistUrl := strings.TrimSpace(string(output))
	modifiedData = fmt.Sprintf("# Source: %s\n%s", gistUrl, modifiedData)
	err = os.WriteFile(gist, []byte(modifiedData), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) Update(repo, title, videoID string) (string, error) {
	if len(title) == 0 {
		errorMessage = "Please set the title of the video first"
		return "", fmt.Errorf(errorMessage)
	}
	if len(videoID) == 0 {
		errorMessage = "Please upload the video first"
		return "", fmt.Errorf(errorMessage)
	}
	repo, err := getInputFromString("What is the name of the repo?", repo)
	if repo == "N/A" {
		return repo, nil
	}
	if err != nil {
		errorMessage = err.Error()
		return "", err
	}
	cmdClone := exec.Command("gh", "repo", "clone", repo)
	output, err := cmdClone.CombinedOutput()
	if err != nil {
		errorMessage = string(output)
		return "", err
	}
	readmePath := fmt.Sprintf("%s/README.md", repo)
	file, err := os.OpenFile(readmePath, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		errorMessage = err.Error()
		return "", err
	}
	contentTitle := "# Demo Manifests and Code Used in DevOps Toolkit Videos"
	contentVideo := fmt.Sprintf("[![%s](https://img.youtube.com/vi/%s/0.jpg)](https://youtu.be/%s)", title, videoID, videoID)
	content := fmt.Sprintf("%s\n\n%s", contentTitle, contentVideo)
	_, err = io.WriteString(file, content)
	if err != nil {
		errorMessage = err.Error()
		return "", err
	}
	cmdGit := []*exec.Cmd{
		exec.Command("git", "add", "."),
		exec.Command("git", "commit", "-m", "Update README"),
		exec.Command("git", "push"),
	}
	for _, cmd := range cmdGit {
		cmd.Dir = repo
		output, err = cmd.CombinedOutput()
		if err != nil {
			errorMessage = string(output)
			return "", err
		}
	}
	cmdDelete := exec.Command("rm", "-rf", repo)
	output, err = cmdDelete.CombinedOutput()
	if err != nil {
		errorMessage = string(output)
		return "", err
	}
	return repo, err
}
