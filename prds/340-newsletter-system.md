# PRD: Newsletter System for YouTube Channel

**Issue**: #340
**Status**: Planning
**Priority**: Medium
**Created**: 2025-11-11
**Last Updated**: 2025-11-11

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
5. Newsletter sent via ESP (Mailchimp, SendGrid, etc.)
6. Newsletter includes video link, AI summary, and CTA
7. Subscribers get value-add content in their inbox
8. System tracks newsletter send status in video YAML

## Success Criteria

### Must Have (MVP)
- [ ] Newsletter automatically sent after configurable delay (settings.yaml)
- [ ] Integration with at least one ESP (to be decided: Mailchimp, SendGrid, or ConvertKit)
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

- Provider implementations (start with one, expand later):
  - `MailchimpProvider` (if Mailchimp chosen)
  - `SendGridProvider` (if SendGrid chosen)
  - `ConvertKitProvider` (if ConvertKit chosen)

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
    provider: "mailchimp"  # or "sendgrid", "convertkit"
    delay_hours: 48
    from_name: "Your Channel Name"
    from_email: "newsletter@yourdomain.com"
    reply_to: "reply@yourdomain.com"

    # Provider-specific config
    mailchimp:
      api_key: "MAILCHIMP_API_KEY"  # from env var
      list_id: "your-list-id"
      template_id: "optional-template-id"

    sendgrid:
      api_key: "SENDGRID_API_KEY"  # from env var
      sender_id: "your-sender-id"

    convertkit:
      api_key: "CONVERTKIT_API_KEY"  # from env var
      form_id: "your-form-id"
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

### 1. ESP Provider Selection
**Options**:
- **Mailchimp**: Most popular, generous free tier (500 subscribers, 1K emails/month), excellent deliverability
- **SendGrid**: Developer-friendly, transactional focus, 100 emails/day free, better for high volume
- **ConvertKit**: Creator-focused, superior automation, but more expensive ($9/mo minimum)

**Recommendation**: Start with **SendGrid** for developer experience and API quality, but build interface to allow switching.

**Discussion Topics**:
- Current subscriber count (impacts free tier feasibility)
- Budget considerations
- Existing ESP accounts or preferences
- Future automation needs

### 2. AI-Generated Content Scope
**Options for Newsletter Content**:
- **Minimal**: Just video title + link with brief AI summary (fastest to implement)
- **Standard**: Summary + 3-5 key takeaways (good balance)
- **Rich**: Summary + takeaways + chapter highlights + curated comments (maximum value)

**Recommendation**: Start with **Standard**, expand to Rich later.

**Discussion Topics**:
- Typical video length and complexity
- How much reading subscribers prefer
- Newsletter frequency (affects acceptable length)

### 3. Subscriber Management Strategy
**Options**:
- **ESP-Only**: All subscriber management in ESP platform (simplest)
- **Hybrid**: Basic tracking in tool, management in ESP
- **Integrated**: Full subscriber CRUD in tool, sync with ESP

**Recommendation**: Start with **ESP-Only**, add tracking later if needed.

**Discussion Topics**:
- Current subscriber acquisition strategy
- Need for subscriber segmentation
- Multi-list management (e.g., different lists per video category)

### 4. Scheduling Architecture
**Options**:
- **Cron Job**: Separate script runs periodically to check eligible videos
- **Built-in Scheduler**: Background goroutine in main process
- **External Queue**: Use job queue system (RabbitMQ, Redis)

**Recommendation**: Start with **Cron Job** for simplicity, migrate to built-in scheduler if always-on API mode used.

**Discussion Topics**:
- Typical video publishing frequency
- CLI vs API mode usage patterns
- Infrastructure preferences

### 5. Newsletter Template Design
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
| Subscriber list growth slow | Medium | Medium | Include signup links in videos; promote newsletter |
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
  - `github.com/mailchimp/mailchimp-marketing-go` (if Mailchimp)
  - `github.com/sendgrid/sendgrid-go` (if SendGrid)
  - ConvertKit Go SDK (if ConvertKit - may need to build wrapper)

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

- [ ] **ESP Provider Selected & Integrated**: Research complete, provider chosen, API connection working
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

### 2025-11-11
- PRD created
- GitHub issue #340 opened
- Initial architecture defined
- Identified key decisions needed for ESP, content scope, and scheduling

---

## Notes

### ESP Provider Comparison Summary

| Feature | Mailchimp | SendGrid | ConvertKit |
|---------|-----------|----------|------------|
| **Free Tier** | 500 subs, 1K emails/mo | 100 emails/day | None ($9/mo min) |
| **Pricing** | $13/mo (500 subs) | $19.95/mo (50K/mo) | $9/mo (300 subs) |
| **Best For** | Beginners, marketing | Developers, scale | Creators, automation |
| **API Quality** | Good | Excellent | Good |
| **Deliverability** | Excellent | Excellent | Very Good |
| **Templates** | Built-in builder | Code-based | Built-in builder |
| **Analytics** | Comprehensive | Good | Comprehensive |
| **Learning Curve** | Low | Medium | Low |

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
