# Scrape Slack Spotify

Scrape all links posted in a Slack channel and add them to a public Spotify playlist.

## Usage

1. Create a Spotify app in https://developer.spotify.com/dashboard
2. Click the "Edit Settings" button and add `http://localhost:10028/callback"` as a _Redirect URIs_.
3. Take note of `Client ID` and `Client secret`
4. Create a Slack Application on your Slack workspace 
5. Give it the `channels:history` and `channels:read` scope
6. Install it to your workspace
7. Take note of the _Bot User OAuth Token_
8. Create a Spotify playlist, right click on it and select _Add to my profile_.
9. Run `go run ./... login` and follow the instructions.
10. Assign the mandatory environment variables:
  - `export SPOTIFY_CLIENT_ID=..the client id from 3`
  - `export SPOTIFY_CLIENT_SECRET=..the client secret from 3`
  - `export SLACK_OAUTH_TOKEN=..the token from 7`
11. Run `go run ./... scrape your_slack_channel your_spotify_username your_playlist_name`
