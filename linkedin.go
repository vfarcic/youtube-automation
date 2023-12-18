package main

import (
	"github.com/atotto/clipboard"
)

func postLinkedIn(message string) {
	clipboard.WriteAll(message)
	println(confirmationStyle.Render("The message has be copied to clipboard. Please paste it into LinkedIn manually."))
}
