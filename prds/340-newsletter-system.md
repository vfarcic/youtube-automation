# PRD: Newsletter System for YouTube Channel

**Issue**: #340
**Status**: Planning
**Priority**: Medium
**Created**: 2025-11-11
**Last Updated**: 2026-02-04

---

## Problem Statement

Currently, there is no automated way to:
- Notify email subscribers when new videos are published
- Provide enhanced content beyond what's available on YouTube
- Capture audience members who prefer email or miss YouTube notifications
- Build a direct relationship with audience independent of platform algorithms

This limits audience reach and engagement, particularly for viewers who:
- Don't enable YouTube notifications
- Miss videos in their crowded subscription feeds
- Prefer digest-style content consumption
- Want direct communication independent of platform changes

## Proposed Solution

Implement an automated newsletter system that:

1. **Sends AI-enhanced video announcements** via email marketing platform
2. **Operates on delayed schedule** (24-48 hours after video publication) to serve as "second wave" promotion
3. **Includes rich content** beyond basic video metadata (AI summaries, key takeaways, etc.)
4. **Integrates with video lifecycle** as part of the publishing workflow
5. **Supports multiple ESP providers** through pluggable architecture
6. **Manages subscribers** through chosen platform's native tools

### Strategic Timing Rationale

**Why 24-48 hour delay?**
- YouTube's algorithm prioritizes first 24 hours - let that work first
- Newsletter becomes **complementary** rather than competitive
- Captures people who missed initial YouTube notification
- Allows including early performance metrics ("Already 50K views!")
- Time to review comments and add community insights
- Reduces "notification fatigue" across multiple channels

### User Journey

**Before (Current State)**:
1. Video published to YouTube
2. YouTube sends notifications to subscribers
3. No email outreach at all
4. Miss potential viewers who prefer email
5. No way to share enhanced context/insights
6. Dependent entirely on platform algorithms

**After (With This Feature)**:
1. Video published to YouTube (existing workflow)
2. System waits configured delay (e.g., 48 hours)
3. AI generates newsletter content (summary, takeaways, etc.)
4. System formats newsletter using template
5. Newsletter sent via ESP (Kit, Mailchimp, etc.)
6. Newsletter includes video link, AI summary, and CTA
7. Subscribers get value-add content in their inbox
8. System tracks newsletter send status in video YAML

## Success Criteria

### Must Have (MVP)
- [ ] Newsletter automatically sent after configurable delay (settings.yaml)
- [ ] Integration with Kit (formerly ConvertKit) as primary ESP
- [ ] AI-generated content included in newsletter (summary, takeaways, or highlights)
- [ ] Newsletter template includes video metadata (title, description, thumbnail, link)
- [ ] Newsletter send status tracked in video YAML
- [ ] Configuration stored in settings.yaml (API keys, delay, ESP settings)
- [ ] Manual "Send Newsletter" option in Publishing Details menu
- [ ] Error handling and retry logic for failed sends

### Nice to Have (Future)
- [ ] Multiple ESP provider support with pluggable architecture
- [ ] A/B testing newsletter subject lines
- [ ] Analytics tracking (open rates, click-through rates)
- [ ] Subscriber management within the tool
- [ ] Newsletter preview/test send before actual send
- [ ] Scheduled digest mode (weekly/monthly roundup of all videos)
- [ ] Personalization (subscriber name, viewing history)
- [ ] Embedded signup forms for website integration

## Technical Scope

### Core Components

#### 1. ESP Integration Module (`internal/platform/newsletter/`)
- Abstract `NewsletterProvider` interface:
  ```go
  type NewsletterProvider interface {
      SendNewsletter(ctx context.Context, content NewsletterContent) error
      GetSubscriberCount(ctx context.Context) (int, error)
      ValidateConnection(ctx context.Context) error
  }

  type NewsletterContent struct {
      Subject     string
      HTMLBody    string
      PlainText   string
      FromName    string
      FromEmail   string
      ReplyTo     string
      VideoID     string  // For tracking
  }
  ```

- Provider implementations (start with Kit, expand later):
  - `KitProvider` (primary - formerly ConvertKit)
  - `MailchimpProvider` (future option)
  - `SendGridProvider` (future option - note: no free tier)

#### 2. Newsletter Content Generator (`internal/newsletter/`)
- `GenerateNewsletterContent()` function
- Composes newsletter from:
  - Video metadata (title, description, link, thumbnail)
  - AI-generated content (summary, takeaways, highlights)
  - Template placeholders
  - Performance metrics (view count, etc.)
- Returns both HTML and plain text versions

#### 3. AI Content Module (`internal/ai/newsletter.go`)
- New AI module for newsletter-specific content
- Generate:
  - Executive summary (3-5 sentences)
  - Key takeaways (bullet points)
  - Chapter highlights (if video has chapters)
  - Engaging subject lines
- Optimized for email reading (concise, scannable)

#### 4. Storage Changes (`internal/storage/yaml.go`)
- Add `Newsletter` field to `Video` struct:
  ```go
  type Video struct {
      // ... existing fields
      Newsletter NewsletterStatus `json:"newsletter,omitempty" yaml:"newsletter,omitempty"`
  }

  type NewsletterStatus struct {
      Sent          bool     `json:"sent" yaml:"sent"`
      SentDate      string   `json:"sent_date,omitempty" yaml:"sent_date,omitempty"`
      ScheduledDate string   `json:"scheduled_date,omitempty" yaml:"scheduled_date,omitempty"`
      Subject       string   `json:"subject,omitempty" yaml:"subject,omitempty"`
      Recipients    int      `json:"recipients,omitempty" yaml:"recipients,omitempty"`
      Error         string   `json:"error,omitempty" yaml:"error,omitempty"`
  }
  ```

#### 5. Configuration (`settings.yaml`)
- Add newsletter section:
  ```yaml
  newsletter:
    enabled: true
    provider: "kit"  # primary choice; alternatives: "mailchimp"
    delay_hours: 48
    from_name: "Your Channel Name"
    from_email: "newsletter@yourdomain.com"
    reply_to: "reply@yourdomain.com"

    # Provider-specific config
    kit:  # formerly ConvertKit - recommended for creators
      api_key: "KIT_API_KEY"  # from env var
      api_secret: "KIT_API_SECRET"  # from env var

    mailchimp:  # alternative - limited free tier (500 contacts, no scheduling)
      api_key: "MAILCHIMP_API_KEY"  # from env var
      list_id: "your-list-id"
      template_id: "optional-template-id"
  ```

#### 6. Scheduling System (`internal/scheduler/`)
- Background process to check for videos needing newsletters
- Query videos where:
  - `Published = true`
  - `Newsletter.Sent = false`
  - `PublishDate + delay_hours < now`
- Process eligible videos and send newsletters
- Update video YAML with send status

#### 7. CLI Interface (`internal/app/`)
- Add "Newsletter" submenu to Publishing Details:
  - "Send Newsletter Now" (manual trigger)
  - "Preview Newsletter" (show what would be sent)
  - "View Newsletter Status" (check if sent, when, stats)
- Add newsletter configuration wizard

#### 8. API Interface (`internal/api/`)
- `GET /videos/{id}/newsletter/preview` - Preview newsletter content
- `POST /videos/{id}/newsletter/send` - Manually trigger newsletter
- `GET /videos/{id}/newsletter/status` - Get newsletter send status
- `GET /newsletter/scheduled` - List videos with pending newsletters
- `POST /newsletter/config/validate` - Test ESP connection

#### 9. Template System
- HTML email template (responsive design)
- Plain text template (fallback)
- Template variables:
  - `{{.Title}}`
  - `{{.Description}}`
  - `{{.VideoURL}}`
  - `{{.ThumbnailURL}}`
  - `{{.AISummary}}`
  - `{{.KeyTakeaways}}`
  - `{{.ViewCount}}`
  - `{{.PublishDate}}`

### Implementation Phases

**Phase 1: ESP Selection & Integration** (Week 1)
- Research and select primary ESP
- Implement provider interface
- Configuration setup
- Connection testing

**Phase 2: Content Generation** (Week 2)
- AI module for newsletter content
- Template system
- HTML/plain text rendering
- Preview functionality

**Phase 3: Storage & Scheduling** (Week 3)
- YAML schema updates
- Scheduling background process
- Status tracking
- Error handling

**Phase 4: CLI & API Interface** (Week 4)
- Publishing Details menu integration
- Manual send controls
- API endpoints
- Documentation

## Open Questions to Discuss During Implementation

### 1. Subscriber Acquisition Strategy ⏳ BLOCKING
**This is the most critical question.** Without a viable subscriber acquisition plan, there is no point building the newsletter system.

**Problem**: Newsletter is only valuable with subscribers. Need to define how to attract and grow the subscriber base before investing in implementation.

**Potential Tactics**:

| Channel | Tactic | Effort | Expected Impact |
|---------|--------|--------|-----------------|
| **In-Video** | Verbal CTA ("link in description") | Low | Medium |
| **In-Video** | End screen overlay with signup link | Low | Medium |
| **In-Video** | Pinned comment with signup link | Low | Low-Medium |
| **Description** | Consistent newsletter section in all videos | Low | Medium |
| **Hugo Blog** | Signup form on blog posts | Medium | Medium |
| **Hugo Blog** | Popup/banner for first-time visitors | Medium | Medium-High |

**Lead Magnet Options** (value exchange for signup):
- **Early access**: Videos 24-48 hours before public release
- **Extended content**: Bonus material cut from videos
- **Resource lists**: Tools, configs, code samples from videos
- **Behind-the-scenes**: Production notes, upcoming content previews

**Best Practices for Tech/DevOps Audience**:
- Exclusive code/configs have highest conversion
- "No spam, just videos" promise reduces friction
- One CTA per video (don't compete with like/subscribe)
- Email-only signup (no name required)

**Discussion Topics**:
- Which lead magnet resonates most with your audience?
- Comfortable with early access model?
- Hugo blog signup form implementation priority?
- Acceptable promotion frequency in videos?

**Go/No-Go Criteria**: Must have a concrete acquisition plan with at least one lead magnet before proceeding with implementation.

---

### 2. ESP Provider Selection ✅ RESOLVED
**Decision**: Use **Kit** (formerly ConvertKit) as the primary ESP.

**Rationale** (decided 2026-02-03):
- **SendGrid eliminated free tier** in May 2025 - no longer viable for cost-free MVP
- **Mailchimp free tier limitations**: 500 contacts (shrinking to 250), no scheduling on free plan, Mailchimp branding required
- **Kit free tier is best**: 10K subscribers, unlimited emails, scheduling supported, creator-focused

**Updated Comparison** (as of 2026):
| Provider | Free Tier | Scheduling | Notes |
|----------|-----------|------------|-------|
| Kit | 10K subs, unlimited emails | ✅ Yes | Best for creators, basic editor |
| Mailchimp | 500 subs (→250), 1K emails/mo | ❌ No | Requires branding, shrinking limits |
| SendGrid | ❌ None (retired May 2025) | N/A | 60-day trial only |

**Known Kit Limitations**:
- Basic email editor (no drag-and-drop)
- Single automation sequence on free tier
- Some deliverability complaints (shared IP issues)
- Price increased Sept 2025 for paid tiers

**Fallback Option**: If Kit deliverability proves problematic, consider **Buttondown** ($9/mo for unlimited) as a developer-friendly alternative.

### 3. AI-Generated Content Scope
**Options for Newsletter Content**:
- **Minimal**: Just video title + link with brief AI summary (fastest to implement)
- **Standard**: Summary + 3-5 key takeaways (good balance)
- **Rich**: Summary + takeaways + chapter highlights + curated comments (maximum value)

**Recommendation**: Start with **Standard**, expand to Rich later.

**Discussion Topics**:
- Typical video length and complexity
- How much reading subscribers prefer
- Newsletter frequency (affects acceptable length)

### 4. Subscriber Management Strategy
**Options**:
- **ESP-Only**: All subscriber management in ESP platform (simplest)
- **Hybrid**: Basic tracking in tool, management in ESP
- **Integrated**: Full subscriber CRUD in tool, sync with ESP

**Recommendation**: Start with **ESP-Only**, add tracking later if needed.

**Discussion Topics**:
- Current subscriber acquisition strategy
- Need for subscriber segmentation
- Multi-list management (e.g., different lists per video category)

### 5. Scheduling Architecture
**Options**:
- **Cron Job**: Separate script runs periodically to check eligible videos
- **Built-in Scheduler**: Background goroutine in main process
- **External Queue**: Use job queue system (RabbitMQ, Redis)

**Recommendation**: Start with **Cron Job** for simplicity, migrate to built-in scheduler if always-on API mode used.

**Discussion Topics**:
- Typical video publishing frequency
- CLI vs API mode usage patterns
- Infrastructure preferences

### 6. Newsletter Template Design
**Options**:
- **ESP Templates**: Use ESP's template builder (less control, easier to maintain)
- **Custom HTML**: Full control, requires frontend development
- **Markdown-based**: Write in markdown, convert to HTML (developer-friendly)

**Recommendation**: **Custom HTML** with inline CSS for maximum control and portability.

**Discussion Topics**:
- Brand guidelines and design preferences
- Need for template variations
- Mobile optimization importance

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Poor email deliverability | High | Medium | Use reputable ESP; implement SPF/DKIM/DMARC |
| Subscriber list growth slow | High | Medium | **BLOCKING**: See Question #1; must resolve before implementation |
| Newsletter content too generic | High | Medium | Iterate on AI prompts; A/B test; gather feedback |
| ESP API rate limits exceeded | High | Low | Implement retry logic; respect rate limits |
| Configuration complexity overwhelms users | Medium | Medium | Provide setup wizard; good defaults |
| Delayed send logic fails | High | Low | Robust scheduling; monitoring; manual fallback |

## Dependencies

### Internal
- Existing AI integration (`internal/ai/`)
- Video YAML storage (`internal/storage/yaml.go`)
- Publishing workflow
- Configuration system

### External
- Email Service Provider (ESP) account and API access
- Azure OpenAI API (existing, for content generation)
- Domain with email sending configured (SPF, DKIM, DMARC)

### New External Dependencies
- ESP SDK/client library:
  - Kit API v4 (REST API - no official Go SDK, will build wrapper)
  - `github.com/mailchimp/mailchimp-marketing-go` (future option)

## Out of Scope

- Custom subscriber management database (use ESP's native tools)
- Advanced automation workflows (drip campaigns, sequences)
- Landing page builder for signup forms
- A/B testing infrastructure (use ESP features if available)
- Analytics dashboard (use ESP analytics)
- Multi-language newsletter support
- RSS feed integration
- Social media post generation from newsletters
- Paid newsletter/subscription features

## Documentation Impact

### New Documentation
- Newsletter setup guide (ESP account, API keys, configuration)
- Newsletter workflow documentation (how/when newsletters are sent)
- Template customization guide
- Subscriber acquisition best practices
- Troubleshooting guide (common ESP issues)

### Updated Documentation
- Publishing workflow (include newsletter step)
- Settings.yaml reference (newsletter section)
- Video YAML schema (newsletter status fields)
- API documentation (newsletter endpoints)
- Getting started guide (newsletter setup as optional step)

## Validation Strategy

### Testing Approach
- **Unit tests**: ESP provider interfaces with mocks
- **Integration tests**: Test sends to sandbox/test accounts
- **End-to-end test**: Full workflow from video publish → newsletter send
- **Template rendering tests**: Validate HTML/plain text output
- **Scheduling tests**: Mock time to verify delay logic

### Manual Validation
- Test with real ESP account (test mode)
- Send to personal email accounts across providers (Gmail, Outlook, etc.)
- Verify mobile rendering (iOS Mail, Gmail app, etc.)
- Check spam folder placement
- Validate link tracking and analytics
- Test error scenarios (invalid API key, network issues)

### Success Metrics (Post-Launch)
- **Technical**: 95%+ newsletter send success rate
- **Engagement**: 20%+ open rate (industry standard for video content)
- **Traffic**: 5%+ click-through rate to videos
- **Growth**: Subscriber count increases over time
- **Reliability**: No missed newsletter sends due to system failure

## Milestones

- [ ] **⚠️ Subscriber Acquisition Plan Approved**: Concrete strategy with lead magnet selected (BLOCKING - must complete before implementation)
- [x] **ESP Provider Selected**: Kit (formerly ConvertKit) chosen based on free tier analysis (2026-02-03)
- [ ] **ESP Provider Integrated**: Kit API connection working
- [ ] **AI Content Generation Working**: Newsletter summaries, takeaways, and subject lines generated successfully
- [ ] **Template System Built**: HTML and plain text templates rendering correctly with all variables
- [ ] **Scheduling System Operational**: Background process reliably identifies and sends newsletters at correct times
- [ ] **Storage & Tracking Complete**: Video YAML properly stores newsletter status, send history, and errors
- [ ] **CLI Workflow Functional**: Users can preview, send, and check newsletter status through Publishing Details
- [ ] **API Endpoints Deployed**: RESTful API supports full newsletter workflow
- [ ] **Documentation Published**: Setup guides, workflow docs, and troubleshooting available
- [ ] **Feature Tested & Validated**: End-to-end testing confirms reliable operation with real ESP
- [ ] **Feature Launched**: Available in production with monitoring and error alerts

## Progress Log

### 2026-02-04
- **Elevated "Subscriber Acquisition Strategy" to Open Question #1 (BLOCKING)**
  - Rationale: Without subscribers, there's no point building the newsletter system
  - Added Go/No-Go criteria: must have concrete acquisition plan before implementation
- Documented potential tactics: in-video CTAs, description links, Hugo blog integration
- Documented lead magnet options: early access, extended content, resource lists
- Added best practices for tech/DevOps audience
- Updated risk table: "Subscriber list growth slow" now marked High impact, BLOCKING

### 2026-02-03
- **ESP Provider Decision**: Selected Kit (formerly ConvertKit) as primary ESP
  - SendGrid no longer viable (free tier retired May 2025)
  - Mailchimp free tier too limited (no scheduling, shrinking limits)
  - Kit offers 10K subscribers free with unlimited emails and scheduling
- **Scope Confirmation**: 500 subscribers sufficient for MVP
- Updated PRD with current ESP pricing/features research
- Added user sentiment analysis from Reddit/Trustpilot reviews
- Identified Kit limitations: basic editor, deliverability concerns, Sept 2025 price hike
- Added Buttondown as fallback option if Kit proves problematic

### 2025-11-11
- PRD created
- GitHub issue #340 opened
- Initial architecture defined
- Identified key decisions needed for ESP, content scope, and scheduling

---

## Notes

### ESP Provider Comparison Summary (Updated 2026-02)

| Feature | Kit (ConvertKit) | Mailchimp | SendGrid |
|---------|------------------|-----------|----------|
| **Free Tier** | ✅ 10K subs, unlimited emails | 500 subs (→250), 1K emails/mo | ❌ None (retired May 2025) |
| **Scheduling** | ✅ Yes | ❌ Not on free | N/A |
| **Best For** | Creators, newsletters | Beginners (small lists) | Developers (paid only) |
| **API Quality** | Good (REST, no Go SDK) | Good | Excellent |
| **Deliverability** | Mixed reviews (shared IP) | Good | Excellent |
| **Templates** | Basic text-focused | Built-in builder | Code-based |
| **Branding** | Minimal | Required on free | N/A |
| **User Sentiment** | Divided (simple vs scale) | "Outgrow it" complaints | Account suspension issues |

**Note**: This comparison was updated after research in Feb 2026. The original Nov 2025 comparison is outdated.

### Content Strategy Recommendations

For maximum newsletter value:
1. **Hook**: Compelling subject line (AI-generated, user can override)
2. **Summary**: 3-5 sentence overview for skimmers
3. **Value-Add**: Key takeaways not obvious from video title
4. **Social Proof**: Early metrics ("10K views in 24 hours!")
5. **CTA**: Clear link to watch video
6. **Bonus**: Link to related videos, blog posts, or resources

### Technical Considerations

- **Rate Limiting**: Most ESPs have rate limits; implement exponential backoff
- **Bounce Handling**: Monitor bounce rates; remove invalid emails
- **Unsubscribe Compliance**: ESP handles this; ensure CAN-SPAM compliance
- **Tracking Pixels**: Use ESP's tracking; don't build custom analytics
- **Deliverability**: Warm up sending domain if using custom domain
