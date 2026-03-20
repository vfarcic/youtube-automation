package notification

import (
	"fmt"
	"html"
	"strings"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"

	gomail "gopkg.in/gomail.v2"
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

func generateThumbnailEmailContent(video storage.Video) (subject, body string) {
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
	subject = fmt.Sprintf("Thumbnail: %s", video.GetUploadTitle())
	taglineIdeas := ""
	if len(video.TaglineIdeas) > 0 && video.TaglineIdeas != "N/A" && video.TaglineIdeas != "-" {
		taglineIdeas = fmt.Sprintf("Ideas:<br/>%s", video.TaglineIdeas)
	}
	body = fmt.Sprintf(`<strong>Material:</strong>
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
	return subject, body
}

func (e *Email) SendThumbnail(from, to string, video storage.Video) error {
	subject, body := generateThumbnailEmailContent(video)
	err := e.Send(from, []string{to}, subject, body, "")
	if err != nil {
		return err
	}
	return nil
}

func generateEditEmailContent(video storage.Video) (subject, body, attachmentPath string, err error) {
	if len(video.Gist) == 0 {
		return "", "", "", fmt.Errorf("Gist is empty")
	}
	subject = fmt.Sprintf("Video: %s", video.GetUploadTitle())
	animations := strings.Split(video.Animations, "\n")
	animationsString := ""
	for i := range animations {
		if strings.TrimSpace(animations[i]) != "" {
			animationsString = fmt.Sprintf("%s\n<li>%s</li>", animationsString, strings.TrimSpace(animations[i]))
		}
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
		video.GetUploadTitle(),
		animationsString,
		video.Members,
	)
	body = fmt.Sprintf(`<strong>Material:</strong>
<br/><br/>
All the material is available at %s.
<br/><br/>
<strong>Animations:</strong>
<ul>
%s
</ul>
`, video.Location, animationsString)
	body = strings.ReplaceAll(body, "\n<li></li>", "")
	if video.AdContent != "" {
		body += fmt.Sprintf(`<br/><br/>
<strong>Sponsor Information:</strong>
<br/><br/>
<pre>%s</pre>
`, html.EscapeString(video.AdContent))
	}
	attachmentPath = video.Gist
	return subject, body, attachmentPath, nil
}

func (e *Email) SendEdit(from, to string, video storage.Video) error {
	subject, body, attachmentPath, err := generateEditEmailContent(video)
	if err != nil {
		return err
	}
	err = e.Send(from, []string{to}, subject, body, attachmentPath)
	if err != nil {
		return err
	}
	return nil
}

func generateSponsorsEmailContent(videoID, sponsorshipPrice, videoTitle string) (subject, body string) {
	subject = fmt.Sprintf("DevOps Toolkit Video Sponsorship - %s", videoTitle)
	body = fmt.Sprintf("Hi,\n<br><br>The video has just been released and is available at https://youtu.be/%s. Please let me know what you think or if you have any questions.\n<br><br>I'll send the invoice for %s in a separate message.\n", videoID, sponsorshipPrice)
	return subject, body
}

func (e *Email) SendSponsors(from, to string, videoID, sponsorshipPrice, videoTitle string) error {
	subject, body := generateSponsorsEmailContent(videoID, sponsorshipPrice, videoTitle)
	toArray := strings.Split(to, ",")
	toArray = append(toArray, configuration.GlobalSettings.Email.FinanceTo)
	err := e.Send(from, toArray, subject, body, "")
	if err != nil {
		return err
	}
	return nil
}

// UploadType distinguishes between a long-form video and a short.
type UploadType string

const (
	UploadTypeVideo UploadType = "Video"
	UploadTypeShort UploadType = "Short"
)

// UploadNotificationParams holds the data needed to build an upload notification email.
type UploadNotificationParams struct {
	Title         string
	Category      string
	Name          string
	YouTubeID     string
	ScheduledDate string
	Type          UploadType
}

func generateUploadNotificationContent(params UploadNotificationParams) (subject, body string) {
	subject = fmt.Sprintf("YouTube Upload: %s (%s)", params.Title, params.Type)

	youtubeURL := fmt.Sprintf("https://youtu.be/%s", params.YouTubeID)
	videoPageURL := fmt.Sprintf("https://youtube.devopstoolkit.ai/videos/%s/%s", params.Category, params.Name)

	scheduledLine := "Not scheduled"
	if params.ScheduledDate != "" {
		scheduledLine = params.ScheduledDate
	}

	body = fmt.Sprintf(`<strong>%s Upload Notification</strong>
<br/><br/>
<ul>
<li><strong>Title:</strong> %s</li>
<li><strong>Type:</strong> %s</li>
<li><strong>Category:</strong> %s</li>
<li><strong>Scheduled Date:</strong> %s</li>
<li><strong>YouTube:</strong> <a href="%s">%s</a></li>
<li><strong>Video Page:</strong> <a href="%s">%s</a></li>
</ul>
`, string(params.Type), params.Title, string(params.Type), params.Category, scheduledLine, youtubeURL, youtubeURL, videoPageURL, videoPageURL)

	return subject, body
}

func (e *Email) SendUploadNotification(from string, params UploadNotificationParams) error {
	subject, body := generateUploadNotificationContent(params)
	return e.Send(from, []string{}, subject, body, "")
}

// DriveUploadNotificationParams holds the data needed to build a Drive upload notification email.
type DriveUploadNotificationParams struct {
	Title       string
	Category    string
	Name        string
	DriveFileID string
	Type        UploadType
}

func generateDriveUploadNotificationContent(params DriveUploadNotificationParams) (subject, body string) {
	subject = fmt.Sprintf("Drive Upload: %s (%s)", params.Title, params.Type)

	videoPageURL := fmt.Sprintf("https://youtube.devopstoolkit.ai/videos/%s/%s", params.Category, params.Name)

	body = fmt.Sprintf(`<strong>%s Drive Upload Notification</strong>
<br/><br/>
<ul>
<li><strong>Title:</strong> %s</li>
<li><strong>Type:</strong> %s</li>
<li><strong>Category:</strong> %s</li>
<li><strong>Drive File ID:</strong> %s</li>
<li><strong>Video Page:</strong> <a href="%s">%s</a></li>
</ul>
`, string(params.Type), params.Title, string(params.Type), params.Category, params.DriveFileID, videoPageURL, videoPageURL)

	return subject, body
}

func (e *Email) SendDriveUploadNotification(from string, params DriveUploadNotificationParams) error {
	subject, body := generateDriveUploadNotificationContent(params)
	return e.Send(from, []string{}, subject, body, "")
}
