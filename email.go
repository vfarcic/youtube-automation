package main

import (
	"fmt"
	"os"
	"strings"

	gomail "gopkg.in/mail.v2"
)

func sendEmail(from string, to []string, subject, body string) error {
	password := os.Getenv("EMAIL_PASSWORD")
	to = append(to, from)
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	dialer := gomail.NewDialer("smtp.gmail.com", 587, from, password)
	if err := dialer.DialAndSend(msg); err != nil {
		errorMessage = err.Error()
		return err
	}
	confirmationMessage = "Email Sent Successfully!"
	return nil
}

func sendThumbnailEmail(from, to string, video Video) error {
	logos := ""
	if video.ProjectURL != "" && video.ProjectURL != "-" && video.ProjectURL != "N/A" {
		logos = video.ProjectURL
	}
	if video.OtherLogos != "" && video.OtherLogos != "-" && video.OtherLogos != "N/A" {
		if len(logos) > 0 {
			logos = fmt.Sprintf("%s, ", logos)
		}
		logos = fmt.Sprintf("%s%s", logos, video.OtherLogos)
	}
	if len(logos) > 0 {
		logos = fmt.Sprintf("<li>Logo: %s</li>", logos)
	}
	subject := fmt.Sprintf("Thumbnail: %s", video.ProjectName)
	taglineIdeas := ""
	if len(video.TaglineIdeas) > 0 && video.TaglineIdeas != "N/A" && video.TaglineIdeas != "-" {
		taglineIdeas = fmt.Sprintf("Ideas:<br/>%s", video.TaglineIdeas)
	}
	body := fmt.Sprintf(`<strong>Material:</strong>
<br/><br/>
All the material is available at %s.
<br/><br/>
<strong>Thumbnail:</strong>
<br/><br/>
Elements:
<ul>
%s
<li>Text: %s</li>
<li>Screenshots: screenshot-*.png</li>
</ul>
%s
`, video.Location, logos, video.Tagline, taglineIdeas)
	err := sendEmail(from, []string{to}, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func sendEditEmail(from, to string, video Video) error {
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
	err := sendEmail(from, []string{to}, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func sendSponsorsEmail(from string, to []string, videoID, sponsorshipPrice string) error {
	subject := "DevOps Toolkit Video Sponsorship"
	body := fmt.Sprintf(`Hi,
<br><br>
The video has just been released and is available at https://youtu.be/%s. Please let me know what you think or if you have any questions.
<br><br>
I'll send the invoice for %s in a separate message.
`, videoID, sponsorshipPrice)
	to = append(to, settings.Email.FinanceTo)
	err := sendEmail(from, to, subject, body)
	if err != nil {
		return err
	}
	return nil
}
