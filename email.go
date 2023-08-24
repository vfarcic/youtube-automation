package main

import (
	"fmt"
	"os"
	"strings"

	gomail "gopkg.in/mail.v2"
)

func sendEmail(from, to string, subject, body string) {
	password := os.Getenv("EMAIL_PASSWORD")
	addresses := []string{from, to}
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", addresses...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	dialer := gomail.NewDialer("smtp.gmail.com", 587, from, password)
	if err := dialer.DialAndSend(msg); err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Email Sent Successfully!")
}

func sendThumbnailEmail(from, to string, video Video) {
	logos := video.ProjectURL
	if video.OtherLogos != "" && video.OtherLogos != "-" && video.OtherLogos != "N/A" {
		logos = fmt.Sprintf("%s, %s", logos, video.OtherLogos)
	}
	subject := fmt.Sprintf("Thumbnail: %s", video.ProjectName)
	body := fmt.Sprintf(`<strong>Material:</strong>
<br/><br/>
All the material is available at %s.
<br/><br/>
<strong>Thumbnail:</strong>
<br/><br/>
Elements:
<ul>
<li>Logo: %s</li>
<li>Text: %s</li>
<li>Screenshots: screenshot-*.png</li>
</ul>
Ideas:
<br/>
%s
`, video.Location, logos, video.Tagline, video.TaglineIdeas)
	sendEmail(from, to, subject, body)
}

func sendEditEmail(from, to string, video Video) {
	subject := fmt.Sprintf("Video: %s", video.ProjectName)
	animations := strings.Split(video.Animations, "\n")
	animationsString := ""
	for i := range animations {
		animationsString = fmt.Sprintf("%s\n<li>%s</li>", animationsString, animations[i])
	}
	animationsString = strings.ReplaceAll(animationsString, "- ", "")
	body := fmt.Sprintf(`<strong>Material:</strong>
<br/><br/>
All the material is available at %s.
<br/><br/>
<strong>Animations:</strong>
<ul>
%s
</ul>
`, video.Location, animationsString)
	body = strings.ReplaceAll(body, "\n<li></li>", "")
	sendEmail(from, to, subject, body)
}