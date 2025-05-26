# LinkedIn Integration (`internal/platform/linkedin`)

This package handles the integration with LinkedIn for posting new YouTube videos.

## Features

* Automated posting of videos to LinkedIn
* Update video metadata with LinkedIn posting status and URL
* Fallback to manual clipboard method when API access is not available

## Configuration

### LinkedIn API Token

The LinkedIn API token must be provided via the `LINKEDIN_ACCESS_TOKEN` environment variable.

```bash
export LINKEDIN_ACCESS_TOKEN="your-linkedin-access-token"
```

For security reasons, the token is not stored in the settings.yaml file.

### Settings.yaml Configuration

The following settings can be configured in the `settings.yaml` file:

```yaml
linkedin:
  apiUrl: "https://api.linkedin.com/v2"  # LinkedIn API base URL
```

## Usage

LinkedIn posting is integrated into the YouTube Automation workflow. When a video is ready to be published, the LinkedIn posting will be handled automatically if the API token is available.

If the token is not available, the posting will fall back to the manual clipboard method, which will copy the message to the clipboard and prompt the user to paste it into LinkedIn manually.

## LinkedIn API Access

To use the LinkedIn API, you need to:

1. Create a LinkedIn Developer Application at https://www.linkedin.com/developers/
2. Configure the necessary permissions (r_liteprofile, w_member_social)
3. Generate an access token with the required scopes
4. Set the access token as the `LINKEDIN_ACCESS_TOKEN` environment variable

## Error Handling

The LinkedIn integration includes error handling for:
* Missing access token (falls back to manual posting)
* API errors (with detailed error messages)
* Invalid video data

When an error occurs with the API, it will fall back to the manual clipboard method.