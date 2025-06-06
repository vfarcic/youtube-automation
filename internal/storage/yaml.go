package storage

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YAML provides methods for reading and writing video data to YAML files.
// IndexPath specifies the path to the index file that lists all videos.
// Ensure all fields that need to be accessed from other packages are exported (start with a capital letter).
type YAML struct {
	IndexPath string
}

// VideoIndex holds basic information about a video, used in the index file.
// Ensure all fields that need to be accessed from other packages are exported.
// Already Exported: Name, Category
// To be Exported: None needed beyond current
// No changes to VideoIndex needed for exportability as fields are already capitalized.
// VideoIndex holds basic information about a video, used in the index file.
// Fields Name and Category are already exported.
// Path, if it were to be added here and used by other packages, would need to be capitalized.
type VideoIndex struct {
	Name     string
	Category string
}

// Video represents all data associated with a video project.
// All fields are already exported.
type Video struct {
	Name                 string
	Index                int
	Path                 string
	Category             string
	ProjectName          string
	ProjectURL           string
	Sponsorship          Sponsorship
	Date                 string
	Delayed              bool
	Screen               bool
	Head                 bool
	Thumbnails           bool
	Diagrams             bool
	Title                string
	Description          string
	Highlight            string
	Tags                 string
	DescriptionTags      string
	Location             string
	Tagline              string
	TaglineIdeas         string
	OtherLogos           string
	Screenshots          bool
	RequestThumbnail     bool
	Thumbnail            string
	Language             string
	Members              string
	Animations           string
	RequestEdit          bool
	Movie                bool
	Timecodes            string
	HugoPath             string
	RelatedVideos        string
	UploadVideo          string
	VideoId              string
	Tweet                string
	LinkedInPosted       bool
	SlackPosted          bool
	HNPosted             bool
	DOTPosted            bool
	BlueSkyPosted        bool
	YouTubeHighlight     bool
	YouTubeComment       bool
	YouTubeCommentReply  bool
	Slides               bool
	GDE                  bool
	Repo                 string
	NotifiedSponsors     bool
	AppliedLanguage      string `yaml:"appliedLanguage,omitempty"`
	AppliedAudioLanguage string `yaml:"appliedAudioLanguage,omitempty"`
	AudioLanguage        string `yaml:"audioLanguage,omitempty"`
	Gist                 string `yaml:"gist,omitempty"`
	Code                 bool   `yaml:"code,omitempty"`
}

// Sponsorship holds details about video sponsorship.
// All fields are already exported.
// Ensure all fields that need to be accessed from other packages are exported.
// Already Exported: Amount, Emails, Blocked
// To be Exported: None needed beyond current
// No changes to Sponsorship struct needed for exportability as fields are already capitalized.
// Sponsorship holds details about video sponsorship.
// Fields Amount, Emails, and Blocked are already exported.
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

func (y *YAML) GetVideo(path string) (Video, error) {
	var video Video
	data, err := os.ReadFile(path)
	if err != nil {
		return video, fmt.Errorf("failed to read video file %s: %w", path, err)
	}
	err = yaml.Unmarshal(data, &video)
	if err != nil {
		return video, fmt.Errorf("failed to unmarshal video data from %s: %w", path, err)
	}
	return video, nil
}

func (y *YAML) WriteVideo(video Video, path string) error {
	data, err := yaml.Marshal(&video)
	if err != nil {
		return fmt.Errorf("failed to marshal video data for %s: %w", path, err)
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write video data to file %s: %w", path, err)
	}
	return nil
}

func (y *YAML) GetIndex() ([]VideoIndex, error) {
	var index []VideoIndex
	data, err := os.ReadFile(y.IndexPath)
	if err != nil {
		return index, fmt.Errorf("failed to read index file %s: %w", y.IndexPath, err)
	}
	err = yaml.Unmarshal(data, &index)
	if err != nil {
		return index, fmt.Errorf("failed to unmarshal video index from %s: %w", y.IndexPath, err)
	}
	return index, nil
}

func (y *YAML) WriteIndex(vi []VideoIndex) error {
	data, err := yaml.Marshal(&vi)
	if err != nil {
		return fmt.Errorf("failed to marshal video index: %w", err)
	}
	err = os.WriteFile(y.IndexPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write video index to file %s: %w", y.IndexPath, err)
	}
	return nil
}
