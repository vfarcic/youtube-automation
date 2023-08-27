package main

import (
	"github.com/atotto/clipboard"
)

func postLinkedIn(message string, posted bool) bool {
	if len(message) == 0 {
		errorMessage = "Please generate Tweet first."
		return false
	}
	if !posted {
		clipboard.WriteAll(message)
		confirmationMessage = "The message has be copied to clipboard. Please paste it into LinkedIn manually."
	}
	return getInputFromBool(posted)
}
