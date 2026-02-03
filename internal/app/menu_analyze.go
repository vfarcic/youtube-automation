package app

import (
	"context"
	"fmt"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/publishing"

	"github.com/charmbracelet/huh"
)

// HandleAnalyzeMenu displays the Analyze submenu with options
func (m *MenuHandler) HandleAnalyzeMenu() error {
	var selectedOption int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("What would you like to analyze?").
				Options(
					huh.NewOption("Titles (fetch video analytics)", 0),
					huh.NewOption("Timing (generate publish time recommendations)", 1),
					huh.NewOption("Sponsor Page (update analytics charts)", 2),
					huh.NewOption("Back", actionReturn),
				).
				Value(&selectedOption),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("failed to run analyze menu form: %w", err)
	}

	switch selectedOption {
	case 0:
		return m.HandleAnalyzeTitles()
	case 1:
		return m.HandleAnalyzeTiming()
	case 2:
		return m.HandleAnalyzeSponsorPage()
	case actionReturn:
		return nil
	}

	return nil
}

// HandleAnalyzeTitles fetches video analytics and displays the results
func (m *MenuHandler) HandleAnalyzeTitles() error {
	fmt.Println(m.normalStyle.Render("Fetching video analytics from YouTube..."))
	fmt.Println(m.normalStyle.Render("This may take a moment and might require re-authentication."))

	ctx := context.Background()
	analytics, err := publishing.GetVideoAnalyticsForLastYear(ctx)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch analytics: %v", err)))
		return err
	}

	if len(analytics) == 0 {
		fmt.Println(m.orangeStyle.Render("No video analytics found for the last 365 days."))
		return nil
	}

	fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ Successfully fetched analytics for %d videos from the last 365 days", len(analytics))))

	// Run AI analysis
	fmt.Println(m.normalStyle.Render("Analyzing title patterns with AI..."))
	fmt.Println(m.normalStyle.Render("This may take a moment."))

	result, prompt, rawResponse, err := ai.AnalyzeTitles(ctx, analytics)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to analyze titles: %v", err)))
		return err
	}

	fmt.Println(m.greenStyle.Render("✓ Analysis complete!"))
	fmt.Println(m.normalStyle.Render("Saving results to files..."))

	// Format result as markdown
	formattedResult := FormatTitleAnalysisMarkdown(result, len(analytics), configuration.GlobalSettings.YouTube.ChannelId)

	// Save complete analysis with all audit trail files
	files, err := SaveCompleteAnalysis(
		"title-analysis",
		analytics,
		prompt,
		rawResponse,
		formattedResult,
		"tmp",
		configuration.GlobalSettings.YouTube.ChannelId,
	)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save files: %v", err)))
		return err
	}

	// Display success message with file paths
	fmt.Println("")
	fmt.Println(m.greenStyle.Render("✓ Analysis files saved successfully!"))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render("Files saved:"))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • Analytics data: %s", files.AnalyticsPath)))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • AI prompt: %s", files.PromptPath)))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • Raw AI response: %s", files.ResponsePath)))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • Formatted analysis: %s", files.ResultPath)))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render("Next steps:"))
	fmt.Println(m.normalStyle.Render("  1. Review the formatted analysis file"))
	fmt.Println(m.normalStyle.Render("  2. Update internal/ai/titles.go with insights"))
	fmt.Println(m.normalStyle.Render("  3. Future titles will use improved patterns"))

	return nil
}

// HandleAnalyzeTiming fetches video analytics and generates timing recommendations
func (m *MenuHandler) HandleAnalyzeTiming() error {
	fmt.Println(m.normalStyle.Render("Fetching video analytics from YouTube..."))
	fmt.Println(m.normalStyle.Render("This may take a moment and might require re-authentication."))

	ctx := context.Background()
	analytics, err := publishing.GetVideoAnalyticsForLastYear(ctx)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch analytics: %v", err)))
		return err
	}

	if len(analytics) == 0 {
		fmt.Println(m.orangeStyle.Render("No video analytics found for the last 365 days."))
		return nil
	}

	fmt.Println(m.greenStyle.Render(fmt.Sprintf("✓ Successfully fetched analytics for %d videos from the last 365 days", len(analytics))))

	// Run AI analysis
	fmt.Println(m.normalStyle.Render("Analyzing timing patterns with AI..."))
	fmt.Println(m.normalStyle.Render("This may take a moment."))

	recommendations, prompt, rawResponse, err := ai.GenerateTimingRecommendations(ctx, analytics)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to generate timing recommendations: %v", err)))
		return err
	}

	fmt.Println(m.greenStyle.Render("✓ Analysis complete!"))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(m.normalStyle.Render("📊 Timing Recommendations"))
	fmt.Println(m.normalStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println("")

	// Display recommendations
	for i, rec := range recommendations {
		fmt.Println(m.greenStyle.Render(fmt.Sprintf("%d. %s %s UTC", i+1, rec.Day, rec.Time)))
		fmt.Println(m.normalStyle.Render(fmt.Sprintf("   %s", rec.Reasoning)))
		fmt.Println("")
	}

	// Ask user if they want to save recommendations
	var saveToSettings bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Save these recommendations to settings.yaml?").
				Description("Recommendations can then be applied to videos using the 'Apply Random Timing' button").
				Affirmative("Yes, save").
				Negative("No, skip").
				Value(&saveToSettings),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to run save confirmation form: %w", err)
	}

	if saveToSettings {
		// Save to settings.yaml
		if err := configuration.SaveTimingRecommendations(recommendations); err != nil {
			fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save to settings.yaml: %v", err)))
			return err
		}
		fmt.Println(m.greenStyle.Render("✓ Recommendations saved to settings.yaml"))
	}

	fmt.Println(m.normalStyle.Render("Saving analysis files..."))

	// Format result as markdown
	formattedResult := FormatTimingRecommendationsMarkdown(recommendations, len(analytics), configuration.GlobalSettings.YouTube.ChannelId)

	// Save complete analysis with all audit trail files
	files, err := SaveCompleteAnalysis(
		"timing-analysis",
		analytics,
		prompt,
		rawResponse,
		formattedResult,
		"tmp",
		configuration.GlobalSettings.YouTube.ChannelId,
	)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to save files: %v", err)))
		return err
	}

	// Display success message with file paths
	fmt.Println("")
	fmt.Println(m.greenStyle.Render("✓ Analysis files saved successfully!"))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render("Files saved:"))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • Analytics data: %s", files.AnalyticsPath)))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • AI prompt: %s", files.PromptPath)))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • Raw AI response: %s", files.ResponsePath)))
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("  • Formatted analysis: %s", files.ResultPath)))
	fmt.Println("")

	if saveToSettings {
		fmt.Println(m.normalStyle.Render("Next steps:"))
		fmt.Println(m.normalStyle.Render("  1. Review the formatted analysis file"))
		fmt.Println(m.normalStyle.Render("  2. Use 'Apply Random Timing' button when editing videos"))
		fmt.Println(m.normalStyle.Render("  3. Re-run analysis in 3-6 months to evolve recommendations"))
	}

	return nil
}

// HandleAnalyzeSponsorPage fetches YouTube analytics and updates the Hugo sponsor page with charts
func (m *MenuHandler) HandleAnalyzeSponsorPage() error {
	fmt.Println(m.normalStyle.Render("Fetching channel analytics from YouTube..."))
	fmt.Println(m.normalStyle.Render("This may take a moment and might require re-authentication."))

	ctx := context.Background()

	// Fetch demographics data
	fmt.Println(m.normalStyle.Render("  Fetching demographics..."))
	demographics, err := publishing.GetChannelDemographics(ctx)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch demographics: %v", err)))
		return err
	}
	fmt.Println(m.greenStyle.Render("  ✓ Demographics data retrieved"))

	// Fetch geographic distribution
	fmt.Println(m.normalStyle.Render("  Fetching geographic distribution..."))
	geography, err := publishing.GetGeographicDistribution(ctx)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch geographic data: %v", err)))
		return err
	}
	fmt.Println(m.greenStyle.Render("  ✓ Geographic data retrieved"))

	// Fetch channel statistics
	fmt.Println(m.normalStyle.Render("  Fetching channel statistics..."))
	stats, err := publishing.GetChannelStatistics(ctx)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch channel statistics: %v", err)))
		return err
	}
	fmt.Println(m.greenStyle.Render("  ✓ Channel statistics retrieved"))

	// Fetch engagement metrics
	fmt.Println(m.normalStyle.Render("  Fetching engagement metrics..."))
	engagement, err := publishing.GetEngagementMetrics(ctx)
	if err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to fetch engagement metrics: %v", err)))
		return err
	}
	fmt.Println(m.greenStyle.Render("  ✓ Engagement metrics retrieved"))

	// Generate the analytics section with Mermaid charts
	fmt.Println(m.normalStyle.Render("Generating Mermaid charts..."))
	section := publishing.GenerateSponsorAnalyticsSection(demographics, geography, stats, engagement)

	// Update the sponsor page
	fmt.Println(m.normalStyle.Render("Updating sponsor page..."))
	if err := publishing.UpdateSponsorPageAnalytics(section); err != nil {
		fmt.Println(m.errorStyle.Render(fmt.Sprintf("Failed to update sponsor page: %v", err)))
		return err
	}

	// Success feedback
	pagePath := publishing.GetSponsorPagePath()
	fmt.Println("")
	fmt.Println(m.greenStyle.Render("✓ Sponsor page updated successfully!"))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render(fmt.Sprintf("Updated file: %s", pagePath)))
	fmt.Println("")
	fmt.Println(m.normalStyle.Render("Next steps:"))
	fmt.Println(m.normalStyle.Render("  1. Review the changes in the Hugo sponsor page"))
	fmt.Println(m.normalStyle.Render("  2. Commit and deploy the Hugo site"))
	fmt.Println(m.normalStyle.Render("  3. Re-run monthly to keep analytics current"))

	return nil
}
