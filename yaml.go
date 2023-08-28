package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type YAML struct{}

type Videos struct {
	ID       int
	Category string
	Subject  string
}

type Video struct {
	PrePublish          Tasks
	Publish             Tasks
	ProjectName         string
	ProjectURL          string
	Sponsored           string
	SponsoredEmails     []string
	Subject             string
	Date                string
	Code                bool
	Screen              bool
	Head                bool
	Thumbnails          bool
	Diagrams            bool
	Title               string
	Description         string
	Tags                string
	DescriptionTags     string
	Location            string
	Tagline             string
	TaglineIdeas        string
	OtherLogos          string
	Screenshots         bool
	RequestThumbnail    bool
	Thumbnail           string
	Members             string
	Animations          string
	RequestEdit         bool
	Movie               bool
	Timecodes           string
	Gist                string
	RelatedVideos       string
	Playlists           []Playlist
	UploadVideo         string
	VideoId             string
	Tweet               string
	TweetPosted         bool
	LinkedInPosted      bool
	SlackPosted         bool
	RedditPosted        bool
	HNPosted            bool
	TCPosted            bool
	YouTubeHighlight    bool
	YouTubeComment      bool
	YouTubeCommentReply bool
	Slides              bool
	GDE                 bool
	RepoReadme          bool
	TwitterSpace        bool
	NotifiedSponsors    bool
}

func (y *YAML) GetVideo(path string) Video {
	var video Video
	data, err := os.ReadFile(path)
	if err != nil {
		return video
	}
	err = yaml.Unmarshal(data, &video)
	if err != nil {
		log.Fatal(err)
	}
	return video
}

func (y *YAML) WriteVideo(video Video, path string) {
	data, err := yaml.Marshal(&video)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
