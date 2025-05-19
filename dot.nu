#!/usr/bin/env nu

source scripts/mcp.nu

def main [] {}

def --env "main setup" [] {

    let anthropic_key_env = ($env | get --ignore-errors ANTHROPIC_API_KEY | default "")
    let anthropic_key = if ($anthropic_key_env | is-not-empty) {
        $anthropic_key_env
    } else {
        input $"(ansi yellow_bold)Anthropic API Key: (ansi reset)"
    }

    let github_token_env = ($env | get --ignore-errors GITHUB_TOKEN | default "")
    let github_token = if ($github_token_env | is-not-empty) {
        $github_token_env
    } else {
        input $"(ansi yellow_bold)GitHub Token: (ansi reset)"
    }

    main apply mcp --anthropic-api-key $anthropic_key --github-token $github_token

    print $"
We're (ansi yellow_bold)not yet done(ansi reset).
Please perform the following actions manually:
1. (ansi yellow_bold)Install Go(ansi reset) \(https://go.dev/doc/install\) if you don't have it already.
2. (ansi yellow_bold)Install Cursor(ansi reset) \(https://www.cursor.com/downloads\) if you don't have it already.
3. If you did NOT already (ansi yellow_bold)set `code` alias(ansi reset), open Cursor, press `cmd+shift+p` and select `Shell Command: Install 'code' command`.
4. (ansi yellow_bold)Close Cursor(ansi reset).
"
    
}
