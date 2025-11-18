# PRD: AI Title Generation Improvement via A/B Test Analysis

| Metadata | Details |
|:---|:---|
| **PRD ID** | 344 |
| **Issue** | [#344](https://github.com/vfarcic/youtube-automation/issues/344) |
| **Feature Name** | AI Title Generation Improvement |
| **Status** | Draft |
| **Priority** | Medium |
| **Author** | @vfarcic |
| **Created** | 2025-11-18 |

## 1. Problem Statement

We are currently collecting A/B test data (Watch Shares) for video titles, but this valuable performance data is not being fed back into our AI title generation process. The AI continues to generate titles based on static templates and generic best practices, missing the opportunity to "learn" from what actually resonates with our specific audience.

## 2. Proposed Solution

Implement a data-driven feedback loop that:
1.  **Analyzes** existing videos to find those with completed A/B tests (where titles have `share > 0`).
2.  **Identifies** "winning" titles (highest share %).
3.  **Extracts** these high-performing titles to serve as dynamic "few-shot" examples or context.
4.  **Updates** the AI prompt context at runtime (or periodically updates the template) to include these real-world success stories.

This will allow the AI to generate new titles that are stylistically similar to our best-performing content.

## 3. User Stories

*   **As a** content creator,
*   **I want** the AI to suggest titles that are proven to work for my channel,
*   **So that** I can increase my Click-Through Rate (CTR) and average view duration without manual guesswork.

## 4. Functional Requirements

### 4.1. Data Extraction
*   Query the video storage (YAML files) for videos with non-empty `titles` arrays.
*   Filter for titles where `share` > 0 (indicating A/B test data exists).
*   Select the title with the highest `share` percentage for each video as the "winner".

### 4.2. Analysis Logic
*   Aggregate winning titles.
*   (Optional) Categorize winners by video category (e.g., "Tutorials", "Opinion", "News") if data volume permits.

### 4.3. Prompt Enhancement
*   Modify the `internal/ai/analyze_titles.go` (or similar logic) to accept a list of high-performing examples.
*   Inject these examples into the system prompt sent to the AI provider.
*   **Format**: "Here are examples of titles that have performed well on this channel: [List of titles]..."

## 5. Technical Implementation

### 5.1. New Components
*   `AnalysisService`: A service in `internal/analysis/` (or similar) responsible for scanning `storage.Video` objects and returning `[]TitleVariant` of winners.

### 5.2. Modified Components
*   `internal/ai/titles.go`: Update `GenerateTitles` function to:
    1.  Call `AnalysisService` to get context.
    2.  Template the context into the prompt before sending to LLM.

## 6. Milestones

- [ ] **Milestone 1: Data Analysis Service**
    - Implement logic to scan all videos and identify "winning" titles based on `share` percentage.
    - Create unit tests with sample video data (some with A/B tests, some without).

- [ ] **Milestone 2: Prompt Integration**
    - Update the title generation prompt template (`internal/ai/templates/titles.md`) to accept a dynamic list of "proven winners".
    - Update `internal/ai` code to fetch winners and inject them into the template.

- [ ] **Milestone 3: End-to-End Testing**
    - Verify that running the title generator now includes the winning titles in the prompt (via logs or debug mode).
    - Ensure fallback behavior works if no A/B test data is available yet.

## 7. Success Criteria

*   **Technical**: The AI prompt sent to the provider includes at least 3-5 examples of high-performing titles from previous videos (when data exists).
*   **Performance**: No significant increase in latency for title generation.
*   **Quality**: Generated titles show stylistic alignment with top-performing past titles.

## 8. Dependencies

*   Existing A/B test data structure (implemented in `storage.Video`).
*   `internal/ai` package for prompt management.

## 9. Risk Assessment

*   **Risk**: Skewed data from older videos might encourage outdated styles.
    *   *Mitigation*: Limit analysis to the last N months or N videos.
*   **Risk**: Prompt context limit.
    *   *Mitigation*: Limit the number of examples injected (e.g., top 5 or top 10).

## 10. Documentation

*   Update `docs/development.md` to explain how the feedback loop works.
