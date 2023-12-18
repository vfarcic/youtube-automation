package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Repo struct{}

func (r *Repo) Gist(gist, title, projectName, projectUrl, relatedVideos string) (string, error) {
	data, err := os.ReadFile(gist)
	if err != nil {
		return "", err
	}
	if gist == "N/A" {
		return "", nil
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
		return "", err
	}
	cmd := exec.Command("gh", "gist", "create", "--public", gist)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	gistUrl := strings.TrimSpace(string(output))
	modifiedData = fmt.Sprintf("# Source: %s\n%s", gistUrl, modifiedData)
	err = os.WriteFile(gist, []byte(modifiedData), 0644)
	if err != nil {
		return "", err
	}
	return gistUrl, nil
}

func (r *Repo) Update(repo, title, videoID string) error {
	cmdClone := exec.Command("gh", "repo", "clone", repo)
	_, err := cmdClone.CombinedOutput()
	if err != nil {
		return err
	}
	readmePath := fmt.Sprintf("%s/README.md", repo)
	file, err := os.OpenFile(readmePath, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		return err
	}
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
