package publishing

import (
	"fmt"
	"strings"
)

// countryCodeToName maps ISO 3166-1 alpha-2 codes to readable country names.
// Limited to top 10 most common countries for YouTube tech channels.
var countryCodeToName = map[string]string{
	"US": "United States",
	"IN": "India",
	"GB": "United Kingdom",
	"DE": "Germany",
	"CA": "Canada",
	"BR": "Brazil",
	"AU": "Australia",
	"NL": "Netherlands",
	"FR": "France",
	"PL": "Poland",
}

// formatAgeGroup converts API age group format to readable format.
// e.g., "age18-24" -> "18-24", "age65-" -> "65+"
func formatAgeGroup(ageGroup string) string {
	// Remove "age" prefix
	formatted := strings.TrimPrefix(ageGroup, "age")
	// Convert trailing dash to plus for 65+
	if strings.HasSuffix(formatted, "-") {
		formatted = strings.TrimSuffix(formatted, "-") + "+"
	}
	return formatted
}

// formatGender converts API gender format to readable format.
// e.g., "male" -> "Male", "female" -> "Female", "user_specified" -> "Other"
func formatGender(gender string) string {
	switch strings.ToLower(gender) {
	case "male":
		return "Male"
	case "female":
		return "Female"
	case "user_specified":
		return "Other"
	default:
		return gender
	}
}

// getCountryName returns the full country name for a code, or the code itself if unknown.
func getCountryName(code string) string {
	if name, ok := countryCodeToName[code]; ok {
		return name
	}
	return code
}

// formatNumber formats large numbers with commas for readability.
// e.g., 1234567 -> "1,234,567"
func formatNumber(n int64) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Build string from right to left
	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// GenerateAgeDistributionChart creates a Mermaid pie chart for age group distribution.
// Returns empty string if no data is available.
func GenerateAgeDistributionChart(demographics ChannelDemographics) string {
	if len(demographics.AgeGroups) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("pie showData title Age Distribution\n")

	for _, ag := range demographics.AgeGroups {
		// Skip very small percentages
		if ag.Percentage < 0.5 {
			continue
		}
		label := formatAgeGroup(ag.AgeGroup)
		sb.WriteString(fmt.Sprintf("    \"%s\" : %.1f\n", label, ag.Percentage))
	}

	sb.WriteString("```")
	return sb.String()
}

// GenerateGenderDistributionChart creates a Mermaid pie chart for gender distribution.
// Returns empty string if no data is available.
func GenerateGenderDistributionChart(demographics ChannelDemographics) string {
	if len(demographics.Gender) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("pie showData title Gender Distribution\n")

	for _, g := range demographics.Gender {
		// Skip very small percentages
		if g.Percentage < 0.5 {
			continue
		}
		label := formatGender(g.Gender)
		sb.WriteString(fmt.Sprintf("    \"%s\" : %.1f\n", label, g.Percentage))
	}

	sb.WriteString("```")
	return sb.String()
}

// GenerateGeographyChart creates a Mermaid bar chart for top countries by views.
// Returns empty string if no data is available.
func GenerateGeographyChart(distribution GeographicDistribution) string {
	if len(distribution.Countries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("xychart-beta horizontal\n")
	sb.WriteString("    title \"Top Countries by Views\"\n")

	// Build x-axis labels and bar values
	var labels []string
	var values []string
	maxPercentage := 0.0

	for _, c := range distribution.Countries {
		labels = append(labels, getCountryName(c.CountryCode))
		values = append(values, fmt.Sprintf("%.1f", c.Percentage))
		if c.Percentage > maxPercentage {
			maxPercentage = c.Percentage
		}
	}

	// Round up max percentage for y-axis
	yAxisMax := int((maxPercentage/10)+1) * 10
	if yAxisMax < 10 {
		yAxisMax = 10
	}

	sb.WriteString(fmt.Sprintf("    x-axis [%s]\n", strings.Join(labels, ", ")))
	sb.WriteString(fmt.Sprintf("    y-axis \"Percentage\" 0 --> %d\n", yAxisMax))
	sb.WriteString(fmt.Sprintf("    bar [%s]\n", strings.Join(values, ", ")))

	sb.WriteString("```")
	return sb.String()
}

// GenerateChannelStatsTable creates a markdown table with channel statistics.
// Handles hidden subscriber count gracefully.
func GenerateChannelStatsTable(stats ChannelStatistics) string {
	var sb strings.Builder
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")

	// Subscriber count (may be hidden)
	if stats.HiddenSubscribers {
		sb.WriteString("| Subscribers | N/A |\n")
	} else {
		sb.WriteString(fmt.Sprintf("| Subscribers | %s |\n", formatNumber(stats.SubscriberCount)))
	}

	sb.WriteString(fmt.Sprintf("| Total Views | %s |\n", formatNumber(stats.TotalViews)))
	sb.WriteString(fmt.Sprintf("| Videos | %s |\n", formatNumber(stats.VideoCount)))

	return sb.String()
}

// GenerateSponsorAnalyticsSection combines all charts into a complete markdown section
// with markers for replacement in the Hugo sponsor page.
func GenerateSponsorAnalyticsSection(demographics ChannelDemographics, distribution GeographicDistribution, stats ChannelStatistics) string {
	var sb strings.Builder

	sb.WriteString("<!-- SPONSOR_ANALYTICS_START -->\n")
	sb.WriteString("## Channel Analytics\n\n")
	sb.WriteString("*Data from the last 90 days, updated automatically.*\n\n")

	// Channel Statistics
	sb.WriteString("### Overview\n\n")
	sb.WriteString(GenerateChannelStatsTable(stats))
	sb.WriteString("\n")

	// Demographics section
	ageChart := GenerateAgeDistributionChart(demographics)
	genderChart := GenerateGenderDistributionChart(demographics)

	if ageChart != "" || genderChart != "" {
		sb.WriteString("### Audience Demographics\n\n")

		if ageChart != "" {
			sb.WriteString(ageChart)
			sb.WriteString("\n\n")
		}

		if genderChart != "" {
			sb.WriteString(genderChart)
			sb.WriteString("\n\n")
		}
	}

	// Geography section
	geoChart := GenerateGeographyChart(distribution)
	if geoChart != "" {
		sb.WriteString("### Geographic Distribution\n\n")
		sb.WriteString(geoChart)
		sb.WriteString("\n")
	}

	sb.WriteString("<!-- SPONSOR_ANALYTICS_END -->")

	return sb.String()
}
