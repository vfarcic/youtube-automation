package configuration

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadTimingRecommendations reads timing recommendations from settings.yaml
// Returns an empty slice if no recommendations exist or if file doesn't exist
func LoadTimingRecommendations() ([]TimingRecommendation, error) {
	yamlFile, err := os.ReadFile("settings.yaml")
	if err != nil {
		// If file doesn't exist, return empty slice (graceful handling)
		if os.IsNotExist(err) {
			return []TimingRecommendation{}, nil
		}
		return nil, fmt.Errorf("failed to read settings.yaml: %w", err)
	}

	var settings Settings
	if err := yaml.Unmarshal(yamlFile, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings.yaml: %w", err)
	}

	// Return empty slice if recommendations is nil
	if settings.Timing.Recommendations == nil {
		return []TimingRecommendation{}, nil
	}

	return settings.Timing.Recommendations, nil
}

// SaveTimingRecommendations writes timing recommendations to settings.yaml
// Preserves all other settings while updating only the timing section
func SaveTimingRecommendations(recommendations []TimingRecommendation) error {
	// Read existing settings
	yamlFile, err := os.ReadFile("settings.yaml")
	if err != nil {
		return fmt.Errorf("failed to read settings.yaml: %w", err)
	}

	var settings Settings
	if err := yaml.Unmarshal(yamlFile, &settings); err != nil {
		return fmt.Errorf("failed to parse settings.yaml: %w", err)
	}

	// Update timing recommendations
	settings.Timing.Recommendations = recommendations

	// Write back to file
	yamlData, err := yaml.Marshal(&settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile("settings.yaml", yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write settings.yaml: %w", err)
	}

	return nil
}
