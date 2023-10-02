package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Repo struct{}

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
