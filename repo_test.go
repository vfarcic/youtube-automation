package main

import (
	"reflect"
	"testing"
)

func TestRepo_getAnimationsFromMarkdown(t *testing.T) {
	repo := &Repo{}
	filePath := "manuscript/category-04/something.md"

	animations, sections, err := repo.getAnimationsFromMarkdown(filePath)
	if err != nil {
		t.Errorf("Error occurred while getting animations and sections: %v", err)
	}

	expectedAnimations := []string{
		"Logo: okteto.png, gitpod.png",
		"Logos: charm-gum.png, git.png, github.png, kind.png, kubectl (there is no logo so put text instead), yq (there is no logo so put text instead), jq.svg, teller.png, aws.png, azure.png, google-cloud.png",
		"Thumbnail: oosQ3z_9UEM",
		"Text: @ZiggleFingers (big)",
		"Logo: nix.png",
		"Section: Ephemeral Shell Environments with Nix",
		"Logo: oh-my-zsh.png",
		"Logo: nix.png",
		"Overlay: screen-01",
		"Logos: charm-gum.png, kind.png, kubectl (there is no logo so put text instead), yq (there is no logo so put text instead), aws.png, azure.png, google-cloud.png",
		"Logo: oh-my-zsh.png",
		"Miki: Ignore the screen between 09:12 and 09:27.",
		"Overlay: nixos-search; Lower-third: https://search.nixos.org",
		"Overlay: nix-store-gc",
		"Section: Nix Pros and Cons",
		"Logos: jenkins.png, github-actions.png, tekton.png, argo-workflows.png",
	}
	expectedSections := []string{
		"Section: Ephemeral Shell Environments with Nix",
		"Section: Nix Pros and Cons",
	}

	if !reflect.DeepEqual(animations, expectedAnimations) {
		t.Errorf("Expected: %v\nGot: %v", expectedAnimations, animations)
	}

	if !reflect.DeepEqual(sections, expectedSections) {
		t.Errorf("Expected: %v\nGot: %v", expectedSections, sections)
	}
}
