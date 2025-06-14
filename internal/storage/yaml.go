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
	Name                 string      `json:"name" completion:"filled_only"`
	Path                 string      `json:"path" completion:"filled_only"`
	Category             string      `json:"category" completion:"filled_only"`
	ProjectName          string      `json:"projectName" completion:"filled_only"`
	ProjectURL           string      `json:"projectURL" completion:"filled_only"`
	Sponsorship          Sponsorship `json:"sponsorship"`
	Date                 string      `json:"date" completion:"filled_only"`
	Delayed              bool        `json:"delayed" completion:"false_only"`
	Screen               bool        `json:"screen" completion:"true_only"`
	Head                 bool        `json:"head" completion:"true_only"`
	Thumbnails           bool        `json:"thumbnails" completion:"true_only"`
	Diagrams             bool        `json:"diagrams" completion:"true_only"`
	Title                string      `json:"title" completion:"filled_only"`
	Description          string      `json:"description" completion:"filled_only"`
	Highlight            string      `json:"highlight" completion:"filled_only"`
	Tags                 string      `json:"tags" completion:"filled_only"`
	DescriptionTags      string      `json:"descriptionTags" completion:"filled_only"`
	Location             string      `json:"location" completion:"filled_only"`
	Tagline              string      `json:"tagline" completion:"filled_only"`
	TaglineIdeas         string      `json:"taglineIdeas" completion:"filled_only"`
	OtherLogos           string      `json:"otherLogos" completion:"filled_only"`
	Screenshots          bool        `json:"screenshots" completion:"true_only"`
	RequestThumbnail     bool        `json:"requestThumbnail" completion:"true_only"`
	Thumbnail            string      `json:"thumbnail" completion:"filled_only"`
	Language             string      `json:"language" completion:"filled_only"`
	Members              string      `json:"members" completion:"filled_only"`
	Animations           string      `json:"animations" completion:"filled_only"`
	RequestEdit          bool        `json:"requestEdit" completion:"true_only"`
	Movie                bool        `json:"movie" completion:"filled_only"`
	Timecodes            string      `json:"timecodes" completion:"no_fixme"`
	HugoPath             string      `json:"hugoPath" completion:"filled_only"`
	RelatedVideos        string      `json:"relatedVideos" completion:"filled_only"`
	UploadVideo          string      `json:"uploadVideo" completion:"filled_only"`
	VideoId              string      `json:"videoId" completion:"filled_only"`
	Tweet                string      `json:"tweet" completion:"filled_only"`
	LinkedInPosted       bool        `json:"linkedInPosted" completion:"true_only"`
	SlackPosted          bool        `json:"slackPosted" completion:"true_only"`
	HNPosted             bool        `json:"hnPosted" completion:"true_only"`
	DOTPosted            bool        `json:"dotPosted" completion:"true_only"`
	BlueSkyPosted        bool        `json:"blueSkyPosted" completion:"true_only"`
	YouTubeHighlight     bool        `json:"youTubeHighlight" completion:"true_only"`
	YouTubeComment       bool        `json:"youTubeComment" completion:"true_only"`
	YouTubeCommentReply  bool        `json:"youTubeCommentReply" completion:"true_only"`
	Slides               bool        `json:"slides" completion:"true_only"`
	GDE                  bool        `json:"gde" completion:"true_only"`
	Repo                 string      `json:"repo" completion:"filled_only"`
	NotifiedSponsors     bool        `json:"notifiedSponsors" completion:"conditional_sponsors"`
	AppliedLanguage      string      `yaml:"appliedLanguage,omitempty" json:"appliedLanguage,omitempty" completion:"filled_only"`
	AppliedAudioLanguage string      `yaml:"appliedAudioLanguage,omitempty" json:"appliedAudioLanguage,omitempty" completion:"filled_only"`
	AudioLanguage        string      `yaml:"audioLanguage,omitempty" json:"audioLanguage,omitempty" completion:"filled_only"`
	Gist                 string      `yaml:"gist,omitempty" json:"gist,omitempty" completion:"filled_only"`
	Code                 bool        `yaml:"code,omitempty" json:"code,omitempty" completion:"true_only"`
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
	Amount  string `json:"amount" completion:"filled_only"`
	Emails  string `json:"emails" completion:"conditional_sponsorship"`
	Blocked string `json:"blocked" completion:"empty_or_filled"`
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
