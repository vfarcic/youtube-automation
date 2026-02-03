package publishing

import (
	"strings"
	"testing"
)

func TestFormatAgeGroup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "age18-24", input: "age18-24", expected: "18-24"},
		{name: "age25-34", input: "age25-34", expected: "25-34"},
		{name: "age35-44", input: "age35-44", expected: "35-44"},
		{name: "age45-54", input: "age45-54", expected: "45-54"},
		{name: "age55-64", input: "age55-64", expected: "55-64"},
		{name: "age65-", input: "age65-", expected: "65+"},
		{name: "age13-17", input: "age13-17", expected: "13-17"},
		{name: "no prefix", input: "25-34", expected: "25-34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAgeGroup(tt.input)
			if result != tt.expected {
				t.Errorf("formatAgeGroup(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatGender(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "male", input: "male", expected: "Male"},
		{name: "female", input: "female", expected: "Female"},
		{name: "user_specified", input: "user_specified", expected: "Other"},
		{name: "MALE uppercase", input: "MALE", expected: "Male"},
		{name: "Female mixed", input: "Female", expected: "Female"},
		{name: "unknown", input: "unknown", expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGender(tt.input)
			if result != tt.expected {
				t.Errorf("formatGender(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetCountryName(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{name: "US", code: "US", expected: "United States"},
		{name: "IN", code: "IN", expected: "India"},
		{name: "GB", code: "GB", expected: "United Kingdom"},
		{name: "DE", code: "DE", expected: "Germany"},
		{name: "CA", code: "CA", expected: "Canada"},
		{name: "BR", code: "BR", expected: "Brazil"},
		{name: "AU", code: "AU", expected: "Australia"},
		{name: "NL", code: "NL", expected: "Netherlands"},
		{name: "FR", code: "FR", expected: "France"},
		{name: "PL", code: "PL", expected: "Poland"},
		{name: "unknown code", code: "ZZ", expected: "ZZ"},
		{name: "JP not in map", code: "JP", expected: "JP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCountryName(tt.code)
			if result != tt.expected {
				t.Errorf("getCountryName(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{name: "zero", input: 0, expected: "0"},
		{name: "small", input: 123, expected: "123"},
		{name: "thousand", input: 1000, expected: "1,000"},
		{name: "tens of thousands", input: 12345, expected: "12,345"},
		{name: "hundreds of thousands", input: 123456, expected: "123,456"},
		{name: "millions", input: 1234567, expected: "1,234,567"},
		{name: "tens of millions", input: 25000000, expected: "25,000,000"},
		{name: "negative", input: -1234, expected: "-1,234"},
		{name: "999", input: 999, expected: "999"},
		{name: "1001", input: 1001, expected: "1,001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("formatNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateAgeDistributionChart(t *testing.T) {
	tests := []struct {
		name         string
		demographics ChannelDemographics
		wantContains []string
		wantEmpty    bool
	}{
		{
			name: "valid demographics",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{
					{AgeGroup: "age18-24", Percentage: 25.5},
					{AgeGroup: "age25-34", Percentage: 42.3},
					{AgeGroup: "age35-44", Percentage: 18.2},
				},
			},
			wantContains: []string{
				"```mermaid",
				"pie showData title Age Distribution",
				`"18-24" : 25.5`,
				`"25-34" : 42.3`,
				`"35-44" : 18.2`,
				"```",
			},
			wantEmpty: false,
		},
		{
			name: "with 65+ age group",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{
					{AgeGroup: "age65-", Percentage: 5.0},
				},
			},
			wantContains: []string{
				`"65+" : 5.0`,
			},
			wantEmpty: false,
		},
		{
			name: "skips small percentages",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{
					{AgeGroup: "age18-24", Percentage: 25.0},
					{AgeGroup: "age13-17", Percentage: 0.3}, // Should be skipped
				},
			},
			wantContains: []string{
				`"18-24" : 25.0`,
			},
			wantEmpty: false,
		},
		{
			name: "empty demographics",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{},
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateAgeDistributionChart(tt.demographics)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}

			// Verify small percentage was skipped
			if tt.name == "skips small percentages" && strings.Contains(result, "13-17") {
				t.Error("result should not contain 13-17 (small percentage)")
			}
		})
	}
}

func TestGenerateGenderDistributionChart(t *testing.T) {
	tests := []struct {
		name         string
		demographics ChannelDemographics
		wantContains []string
		wantEmpty    bool
	}{
		{
			name: "valid gender data",
			demographics: ChannelDemographics{
				Gender: []GenderData{
					{Gender: "male", Percentage: 85.2},
					{Gender: "female", Percentage: 14.5},
				},
			},
			wantContains: []string{
				"```mermaid",
				"pie showData title Gender Distribution",
				`"Male" : 85.2`,
				`"Female" : 14.5`,
				"```",
			},
			wantEmpty: false,
		},
		{
			name: "with user_specified",
			demographics: ChannelDemographics{
				Gender: []GenderData{
					{Gender: "male", Percentage: 80.0},
					{Gender: "female", Percentage: 15.0},
					{Gender: "user_specified", Percentage: 5.0},
				},
			},
			wantContains: []string{
				`"Other" : 5.0`,
			},
			wantEmpty: false,
		},
		{
			name: "skips small percentages",
			demographics: ChannelDemographics{
				Gender: []GenderData{
					{Gender: "male", Percentage: 85.0},
					{Gender: "user_specified", Percentage: 0.2}, // Should be skipped
				},
			},
			wantContains: []string{
				`"Male" : 85.0`,
			},
			wantEmpty: false,
		},
		{
			name: "empty gender data",
			demographics: ChannelDemographics{
				Gender: []GenderData{},
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateGenderDistributionChart(tt.demographics)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}

			// Verify small percentage was skipped
			if tt.name == "skips small percentages" && strings.Contains(result, "Other") {
				t.Error("result should not contain Other (small percentage)")
			}
		})
	}
}

func TestGenerateGeographyChart(t *testing.T) {
	tests := []struct {
		name         string
		distribution GeographicDistribution
		wantContains []string
		wantEmpty    bool
	}{
		{
			name: "valid country data",
			distribution: GeographicDistribution{
				Countries: []CountryData{
					{CountryCode: "US", Views: 35000, Percentage: 35.0},
					{CountryCode: "IN", Views: 12500, Percentage: 12.5},
					{CountryCode: "GB", Views: 8300, Percentage: 8.3},
				},
			},
			wantContains: []string{
				"```mermaid",
				"xychart-beta horizontal",
				`title "Top Countries by Views"`,
				"x-axis [United States, India, United Kingdom]",
				"y-axis \"Percentage\" 0 --> 40",
				"bar [35.0, 12.5, 8.3]",
				"```",
			},
			wantEmpty: false,
		},
		{
			name: "unknown country code",
			distribution: GeographicDistribution{
				Countries: []CountryData{
					{CountryCode: "JP", Views: 5000, Percentage: 5.0},
				},
			},
			wantContains: []string{
				"x-axis [JP]",
			},
			wantEmpty: false,
		},
		{
			name: "small percentages round up y-axis",
			distribution: GeographicDistribution{
				Countries: []CountryData{
					{CountryCode: "US", Views: 500, Percentage: 5.0},
				},
			},
			wantContains: []string{
				"y-axis \"Percentage\" 0 --> 10",
			},
			wantEmpty: false,
		},
		{
			name: "empty distribution",
			distribution: GeographicDistribution{
				Countries: []CountryData{},
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateGeographyChart(tt.distribution)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}
		})
	}
}

func TestGenerateChannelStatsTable(t *testing.T) {
	tests := []struct {
		name         string
		stats        ChannelStatistics
		wantContains []string
		wantExcludes []string
	}{
		{
			name: "normal stats",
			stats: ChannelStatistics{
				SubscriberCount:   150000,
				TotalViews:        25000000,
				VideoCount:        450,
				HiddenSubscribers: false,
			},
			wantContains: []string{
				"| Metric | Value |",
				"|--------|-------|",
				"| Subscribers | 150,000 |",
				"| Total Views | 25,000,000 |",
				"| Videos | 450 |",
				"| Avg Views/Video | 55,555 |",
			},
		},
		{
			name: "hidden subscribers",
			stats: ChannelStatistics{
				SubscriberCount:   0,
				TotalViews:        10000000,
				VideoCount:        200,
				HiddenSubscribers: true,
			},
			wantContains: []string{
				"| Subscribers | N/A |",
				"| Total Views | 10,000,000 |",
				"| Videos | 200 |",
			},
			wantExcludes: []string{
				"| Subscribers | 0 |",
			},
		},
		{
			name: "small numbers",
			stats: ChannelStatistics{
				SubscriberCount:   100,
				TotalViews:        5000,
				VideoCount:        10,
				HiddenSubscribers: false,
			},
			wantContains: []string{
				"| Subscribers | 100 |",
				"| Total Views | 5,000 |",
				"| Videos | 10 |",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateChannelStatsTable(tt.stats)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}

			for _, exclude := range tt.wantExcludes {
				if strings.Contains(result, exclude) {
					t.Errorf("result should not contain %q, got:\n%s", exclude, result)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{name: "zero seconds", seconds: 0, expected: "0s"},
		{name: "only seconds", seconds: 45, expected: "45s"},
		{name: "one minute", seconds: 60, expected: "1m 0s"},
		{name: "minutes and seconds", seconds: 330, expected: "5m 30s"},
		{name: "many minutes", seconds: 754, expected: "12m 34s"},
		{name: "fractional seconds", seconds: 123.7, expected: "2m 3s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestGenerateEngagementTable(t *testing.T) {
	tests := []struct {
		name         string
		metrics      EngagementMetrics
		wantContains []string
		wantExcludes []string
	}{
		{
			name: "full engagement data",
			metrics: EngagementMetrics{
				AverageViewDuration: 330.5,
				Likes:               50000,
				Comments:            5000,
				Shares:              2000,
				Views:               1000000,
			},
			wantContains: []string{
				"| Metric | Value |",
				"| Avg Watch Time | 5m 30s |",
				"| Likes | 50,000 |",
				"| Comments | 5,000 |",
				"| Shares | 2,000 |",
				"| Engagement Rate | 5.50% |",
			},
		},
		{
			name: "no views - no engagement rate",
			metrics: EngagementMetrics{
				AverageViewDuration: 120,
				Likes:               1000,
				Comments:            100,
				Shares:              50,
				Views:               0,
			},
			wantContains: []string{
				"| Avg Watch Time | 2m 0s |",
				"| Likes | 1,000 |",
			},
			wantExcludes: []string{
				"| Engagement Rate |",
			},
		},
		{
			name: "small numbers",
			metrics: EngagementMetrics{
				AverageViewDuration: 45,
				Likes:               100,
				Comments:            10,
				Shares:              5,
				Views:               1000,
			},
			wantContains: []string{
				"| Avg Watch Time | 45s |",
				"| Likes | 100 |",
				"| Engagement Rate | 11.00% |",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateEngagementTable(tt.metrics)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}

			for _, exclude := range tt.wantExcludes {
				if strings.Contains(result, exclude) {
					t.Errorf("result should not contain %q, got:\n%s", exclude, result)
				}
			}
		})
	}
}

func TestGenerateSponsorAnalyticsSection(t *testing.T) {
	tests := []struct {
		name         string
		demographics ChannelDemographics
		distribution GeographicDistribution
		stats        ChannelStatistics
		engagement   EngagementMetrics
		wantContains []string
	}{
		{
			name: "full section with all data",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{
					{AgeGroup: "age25-34", Percentage: 40.0},
				},
				Gender: []GenderData{
					{Gender: "male", Percentage: 85.0},
				},
			},
			distribution: GeographicDistribution{
				Countries: []CountryData{
					{CountryCode: "US", Views: 10000, Percentage: 35.0},
				},
			},
			stats: ChannelStatistics{
				SubscriberCount: 100000,
				TotalViews:      5000000,
				VideoCount:      300,
			},
			engagement: EngagementMetrics{
				AverageViewDuration: 300,
				Likes:               50000,
				Comments:            5000,
				Shares:              2000,
				Views:               1000000,
			},
			wantContains: []string{
				"<!-- SPONSOR_ANALYTICS_START -->",
				"## Channel Analytics",
				"*Last updated:",
				"Data covers regular videos from the preceding 90 days (excludes Shorts).*",
				"### Overview",
				"| Subscribers | 100,000 |",
				"### Audience Demographics",
				"pie showData title Age Distribution",
				"pie showData title Gender Distribution",
				"### Geographic Distribution",
				"xychart-beta horizontal",
				"### Engagement",
				"| Avg Watch Time |",
				"| Likes |",
				"<!-- SPONSOR_ANALYTICS_END -->",
			},
		},
		{
			name: "section without demographics",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{},
				Gender:    []GenderData{},
			},
			distribution: GeographicDistribution{
				Countries: []CountryData{
					{CountryCode: "US", Views: 10000, Percentage: 35.0},
				},
			},
			stats: ChannelStatistics{
				SubscriberCount: 50000,
				TotalViews:      2000000,
				VideoCount:      100,
			},
			engagement: EngagementMetrics{},
			wantContains: []string{
				"<!-- SPONSOR_ANALYTICS_START -->",
				"### Overview",
				"### Geographic Distribution",
				"<!-- SPONSOR_ANALYTICS_END -->",
			},
		},
		{
			name: "section without geography",
			demographics: ChannelDemographics{
				AgeGroups: []AgeGroupData{
					{AgeGroup: "age25-34", Percentage: 40.0},
				},
				Gender: []GenderData{},
			},
			distribution: GeographicDistribution{
				Countries: []CountryData{},
			},
			stats: ChannelStatistics{
				SubscriberCount: 50000,
				TotalViews:      2000000,
				VideoCount:      100,
			},
			engagement: EngagementMetrics{},
			wantContains: []string{
				"<!-- SPONSOR_ANALYTICS_START -->",
				"### Overview",
				"### Audience Demographics",
				"<!-- SPONSOR_ANALYTICS_END -->",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSponsorAnalyticsSection(tt.demographics, tt.distribution, tt.stats, tt.engagement)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}

			// Verify sections are properly excluded when data is empty
			if tt.name == "section without demographics" && strings.Contains(result, "### Audience Demographics") {
				t.Error("result should not contain Audience Demographics section when no data")
			}
			if tt.name == "section without geography" && strings.Contains(result, "### Geographic Distribution") {
				t.Error("result should not contain Geographic Distribution section when no data")
			}
		})
	}
}
