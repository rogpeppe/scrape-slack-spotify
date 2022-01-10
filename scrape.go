package main

import (
	"log"
	"regexp"

	"github.com/nlopes/slack"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify"
	"gopkg.in/errgo.v2/errors"
)

var spotifyURIPattern = regexp.MustCompile(`<(?:spotify:track:([a-zA-Z0-9]+)|https://open.spotify.com/track/([a-zA-Z0-9]+)(\?.*)?)>`)

var scrapeCmd = &cobra.Command{
	Use:  "scrape channel destination-playlist",
	RunE: scrape,
	Args: cobra.ExactArgs(2),
}

func scrape(cmd *cobra.Command, args []string) error {
	channelName := args[0]
	playlistName := args[1]
	r, err := client.GetPlaylistsForUser("rogpeppe")
	if err != nil {
		return errors.Note(err, nil, "cannot search")
	}
	var plid spotify.ID
	for _, pl := range r.Playlists {
		if pl.Name == playlistName {
			plid = pl.ID
			break
		}
	}
	if plid == "" {
		return errors.New("could not find heetch playlist")
	}

	if slackOAuthToken == "" {
		return errors.New("no Slack OAuth token found in $SLACK_OAUTH_TOKEN")
	}
	slackAPI := slack.New(slackOAuthToken)

	musicChID, err := getSlackChannelID(slackAPI, channelName)
	if err != nil {
		return errors.Wrap(err)
	}

	tracks := make(chan string, 100)
	go findSpotifyTracks(slackAPI, musicChID, tracks)

	var ids []spotify.ID
	flush := func() error {
		_, err := client.AddTracksToPlaylist(plid, ids...)
		if err != nil {
			return errors.Note(err, nil, "cannot add tracks")
		}
		log.Printf("added %d tracks", len(ids))
		return nil
	}
	for track := range tracks {
		ids = append(ids, spotify.ID(track))
		if len(ids) >= 100 {
			if err := flush(); err != nil {
				return errors.Wrap(err)
			}
			ids = ids[:0]
		}
	}
	if len(ids) > 0 {
		if err := flush(); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func findSpotifyTracks(slackAPI *slack.Client, musicChID string, tracks chan<- string) {
	defer close(tracks)
	var next string
	for {
		hist, err := slackAPI.GetConversationHistory(&slack.GetConversationHistoryParameters{
			Cursor:    next,
			ChannelID: musicChID,
		})
		if err != nil {
			log.Printf("cannot get conversation history: %v", err)
			return
		}
		for _, m := range hist.Messages {
			ma := spotifyURIPattern.FindStringSubmatch(m.Text)
			if len(ma) == 0 {
				continue
			}
			id := ma[1]
			if id == "" {
				id = ma[2]
			}
			tracks <- id
		}
		if hist.ResponseMetaData.NextCursor == "" {
			break
		}
		next = hist.ResponseMetaData.NextCursor
	}
}

func getSlackChannelID(slackAPI *slack.Client, channelName string) (string, error) {
	var next string
	for {
		channels, next1, err := slackAPI.GetConversations(&slack.GetConversationsParameters{
			Cursor: next,
		})
		if err != nil {
			return "", errors.Note(err, nil, "getConversations")
		}
		for _, c := range channels {
			if c.Name == channelName {
				return c.ID, nil
			}
		}
		if next1 == "" {
			break
		}
		next = next1
	}
	return "", errors.New("cannot find channel")
}

//
//func (api *Client) GetConversations(params *GetConversationsParameters) (channels []Channel, nextCursor string, err error)
//
//type GetConversationHistoryParameters struct {
//	ChannelID string
//	Cursor    string
//	Inclusive bool
//	Latest    string
//	Limit     int
//	Oldest    string
//}
//
//type GetConversationHistoryResponse struct {
//	SlackResponse
//	HasMore          bool   `json:"has_more"`
//	PinCount         int    `json:"pin_count"`
//	Latest           string `json:"latest"`
//	ResponseMetaData struct {
//		NextCursor string `json:"next_cursor"`
//	} `json:"response_metadata"`
//	Messages []Message `json:"messages"`
//}
//
//type Message struct {
//	Msg
//	SubMessage      *Msg `json:"message,omitempty"`
//	PreviousMessage *Msg `json:"previous_message,omitempty"`
//}
//type Msg struct {
//	// Basic Message
//	Type            string       `json:"type,omitempty"`
//	Channel         string       `json:"channel,omitempty"`
//	User            string       `json:"user,omitempty"`
//	Text            string       `json:"text,omitempty"`
//	Timestamp       string       `json:"ts,omitempty"`
//	ThreadTimestamp string       `json:"thread_ts,omitempty"`
//	IsStarred       bool         `json:"is_starred,omitempty"`
//	PinnedTo        []string     `json:"pinned_to,omitempty"`
//	Attachments     []Attachment `json:"attachments,omitempty"`
//	Edited          *Edited      `json:"edited,omitempty"`
//	LastRead        string       `json:"last_read,omitempty"`
//	Subscribed      bool         `json:"subscribed,omitempty"`
//	UnreadCount     int          `json:"unread_count,omitempty"`
//
//	// Message Subtypes
//	SubType string `json:"subtype,omitempty"`
//
//	// Hidden Subtypes
//	Hidden           bool   `json:"hidden,omitempty"`     // message_changed, message_deleted, unpinned_item
//	DeletedTimestamp string `json:"deleted_ts,omitempty"` // message_deleted
//	EventTimestamp   string `json:"event_ts,omitempty"`
//
//	// bot_message (https://api.slack.com/events/message/bot_message)
//	BotID    string `json:"bot_id,omitempty"`
//	Username string `json:"username,omitempty"`
//	Icons    *Icon  `json:"icons,omitempty"`
//
//	// channel_join, group_join
//	Inviter string `json:"inviter,omitempty"`
//
//	// channel_topic, group_topic
//	Topic string `json:"topic,omitempty"`
//
//	// channel_purpose, group_purpose
//	Purpose string `json:"purpose,omitempty"`
//
//	// channel_name, group_name
//	Name    string `json:"name,omitempty"`
//	OldName string `json:"old_name,omitempty"`
//
//	// channel_archive, group_archive
//	Members []string `json:"members,omitempty"`
//
//	// channels.replies, groups.replies, im.replies, mpim.replies
//	ReplyCount   int     `json:"reply_count,omitempty"`
//	Replies      []Reply `json:"replies,omitempty"`
//	ParentUserId string  `json:"parent_user_id,omitempty"`
//
//	// file_share, file_comment, file_mention
//	Files []File `json:"files,omitempty"`
//
//	// file_share
//	Upload bool `json:"upload,omitempty"`
//
//	// file_comment
//	Comment *Comment `json:"comment,omitempty"`
//
//	// pinned_item
//	ItemType string `json:"item_type,omitempty"`
//
//	// https://api.slack.com/rtm
//	ReplyTo int    `json:"reply_to,omitempty"`
//	Team    string `json:"team,omitempty"`
//
//	// reactions
//	Reactions []ItemReaction `json:"reactions,omitempty"`
//
//	// slash commands and interactive messages
//	ResponseType    string `json:"response_type,omitempty"`
//	ReplaceOriginal bool   `json:"replace_original"`
//	DeleteOriginal  bool   `json:"delete_original"`
//
//	// Block type Message
//	Blocks Blocks `json:"blocks,omitempty"`
//}
