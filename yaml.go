package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type YAML struct {
	IndexPath string
}

type VideoIndex struct {
	Name     string
	Category string
}

type Video struct {
	Name                string
	Index               int
	Path                string
	Category            string
	Init                Tasks
	Work                Tasks
	Define              Tasks
	Edit                Tasks
	Publish             Tasks
	ProjectName         string
	ProjectURL          string
	Sponsorship         Sponsorship
	Date                string
	Delayed             bool
	Code                bool
	Screen              bool
	Head                bool
	Thumbnails          bool
	Diagrams            bool
	Title               string
	Description         string
	Highlight           string
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
	HugoPath            string
	RelatedVideos       string
	UploadVideo         string
	VideoId             string
	Tweet               string
	LinkedInPosted      bool
	SlackPosted         bool
	HNPosted            bool
	DOTPosted           bool
	BlueSkyPosted       bool
	YouTubeHighlight    bool
	YouTubeComment      bool
	YouTubeCommentReply bool
	Slides              bool
	GDE                 bool
	Repo                string
	NotifiedSponsors    bool
}

type Tasks struct {
	Completed int
	Total     int
}

type Sponsorship struct {
	Amount  string
	Emails  string
	Blocked string
}

// NewYAML creates a new YAML instance with default values
func NewYAML(indexPath string) *YAML {
	return &YAML{
		IndexPath: indexPath,
	}
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

func (y *YAML) GetIndex() []VideoIndex {
	var index []VideoIndex
	data, err := os.ReadFile(y.IndexPath)
	if err != nil {
		return index
	}
	err = yaml.Unmarshal(data, &index)
	if err != nil {
		log.Fatal(err)
	}
	return index
}

func (y *YAML) WriteIndex(vi []VideoIndex) {
	data, err := yaml.Marshal(&vi)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(y.IndexPath, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
