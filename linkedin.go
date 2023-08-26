package main

import (
	"github.com/atotto/clipboard"
)

func postLinkedIn(message string, posted bool) bool {
	if len(message) == 0 {
		println(redStyle.Render("\nPlease generate Tweet first."))
		return false
	}
	if !posted {
		clipboard.WriteAll(message)
		println(orangeStyle.Render("\nThe message has be copied to clipboard. Please paste it into LinkedIn manually."))
	}
	return getInputFromBool(posted)
}
