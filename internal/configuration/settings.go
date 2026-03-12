package configuration

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadTimingRecommendations reads timing recommendations from the given settings file path.
// Returns an empty slice if no recommendations exist or if file doesn't exist.
func LoadTimingRecommendations(settingsPath string) ([]TimingRecommendation, error) {
	yamlFile, err := os.ReadFile(settingsPath)
	if err != nil {
		// If file doesn't exist, return empty slice (graceful handling)
		if os.IsNotExist(err) {
			return []TimingRecommendation{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", settingsPath, err)
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

// SaveTimingRecommendations writes timing recommendations to the given settings file path.
// Preserves all other settings while updating only the timing section.
func SaveTimingRecommendations(settingsPath string, recommendations []TimingRecommendation) error {
	// Read existing settings
	yamlFile, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", settingsPath, err)
	}

	var settings Settings
	if err := yaml.Unmarshal(yamlFile, &settings); err != nil {
		return fmt.Errorf("failed to parse %s: %w", settingsPath, err)
	}

	// Update timing recommendations
	settings.Timing.Recommendations = recommendations

	// Write back to file
	yamlData, err := yaml.Marshal(&settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", settingsPath, err)
	}

	return nil
}
