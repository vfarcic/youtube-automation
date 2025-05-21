package notification

import (
	"fmt"
	"strings"

	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/internal/storage"

	gomail "gopkg.in/mail.v2"
)

type Email struct {
	password string
}

func NewEmail(password string) *Email {
	return &Email{
		password: password,
	}
}

func (e *Email) Send(from string, to []string, subject, body string, attachmentPath string) error {
	to = append(to, from)
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	if attachmentPath != "" {
		msg.Attach(attachmentPath)
	}
	dialer := gomail.NewDialer("smtp.gmail.com", 587, from, e.password)
	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}

func (e *Email) SendThumbnail(from, to string, video storage.Video) error {
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
	err := e.Send(from, []string{to}, subject, body, "")
	if err != nil {
		return err
	}
	return nil
}

func (e *Email) SendEdit(from, to string, video storage.Video) error {
	if len(video.Gist) == 0 {
		return fmt.Errorf("Gist is empty")
	}
	subject := fmt.Sprintf("Video: %s", video.ProjectName)
	animations := strings.Split(video.Animations, "\n")
	animationsString := ""
	for i := range animations {
		animationsString = fmt.Sprintf("%s\n<li>%s</li>", animationsString, animations[i])
	}
	animationsString = strings.ReplaceAll(animationsString, "- ", "")
	animationsString = fmt.Sprintf(`<li>Animation: Subscribe (anywhere in the video)</li>
<li>Animation: Like (anywhere in the video)</li>
<li>Lower third: Viktor Farcic (anywhere in the video)</li>
<li>Animation: Join the channel (anywhere in the video)</li>
<li>Animation: Sponsor the channel (anywhere in the video)</li>
<li>Lower third: %s + logo + URL (%s) (add to a few places when I mention %s)</li>
<li>Text: Transcript and commands + an arrow pointing below (add shortly after we start showing the code)</li>
<li>Title roll: %s</li>
<li>Convert all text in bold (surounded with **) in the attachment into text on the screen</li>
<li>Convert all text in italic (surounded with *) in the attachment into "special" part of the video since those are side-notes.</li>
%s
<li>Member shoutouts: Thanks a ton to the new members for supporting the channel: %s</li>
<li>Outro roll</li>`,
		video.ProjectName,
		video.ProjectURL,
		video.ProjectName,
		video.Title,
		animationsString,
		video.Members,
	)
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
	err := e.Send(from, []string{to}, subject, body, video.Gist)
	if err != nil {
		return err
	}
	return nil
}

func (e *Email) SendSponsors(from, to string, videoID, sponsorshipPrice string) error {
	subject := "DevOps Toolkit Video Sponsorship"
	body := fmt.Sprintf(`Hi,
<br><br>
The video has just been released and is available at https://youtu.be/%s. Please let me know what you think or if you have any questions.
<br><br>
I\'ll send the invoice for %s in a separate message.
`, videoID, sponsorshipPrice)
	toArray := strings.Split(to, ",")
	toArray = append(toArray, configuration.GlobalSettings.Email.FinanceTo)
	err := e.Send(from, toArray, subject, body, "")
	if err != nil {
		return err
	}
	return nil
}
