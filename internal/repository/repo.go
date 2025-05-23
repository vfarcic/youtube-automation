package repository

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"devopstoolkit/youtube-automation/pkg/utils"
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
	// Confirmation before deleting the repository
	confirmMsg := fmt.Sprintf("Are you sure you want to delete the local repository clone at '%s'?", repo)
	if utils.ConfirmAction(confirmMsg, nil) {
		cmdDelete := exec.Command("rm", "-rf", repo)
		_, err = cmdDelete.CombinedOutput()
		if err != nil {
			return err // Return error from rm -rf
		}
		fmt.Printf("Local repository clone '%s' deleted successfully.\n", repo) // Optional: success message
	} else {
		fmt.Printf("Deletion of local repository clone '%s' cancelled.\n", repo)
		// Optionally, return a specific error or nil if cancellation is not an error for the caller
		// For now, if cancelled, we don't return an error, implying the Update was "successful" in not deleting.
	}
	return nil // Ensure a return path if confirmation is no, or if rm -rf succeeds. Original code had `return err` which would be from the last git command if rm -rf wasn't run or succeeded.
}

// GetAnimations extracts animation cues and section titles from the specified markdown file.
// It processes the file line by line:
//   - Lines starting with "TODO:" are considered animation cues; the text after "TODO:" (trimmed) is added to the animations list.
//   - Lines starting with "## " are considered section headers, unless they are "## Intro", "## Setup", or "## Destroy".
//     The text after "## " (trimmed), prefixed with "Section: ", is added to both the animations and sections lists.
//
// It returns a slice of animation strings, a slice of section title strings, and any error encountered.
func (r *Repo) GetAnimations(filePath string) (animations, sections []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, " ", " ") // Non-breaking space
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

	if animations == nil {
		animations = []string{}
	}
	if sections == nil {
		sections = []string{}
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
