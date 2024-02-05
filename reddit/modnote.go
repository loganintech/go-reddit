package reddit

import (
	"context"
	"net/http"
	"strings"
)

type ModActionData struct {
	Action      *string `json:"action"`
	RedditId    *string `json:"reddit_id"`
	Details     *string `json:"details"`
	Description *string `json:"description"`
}

type UserNoteData struct {
	Note     *string `json:"note"`
	RedditId *string `json:"reddit_id"`
	Label    *string `json:"label"`
}

type Modnote struct {
	SubredditId   string        `json:"subreddit_id"`
	OperatorId    string        `json:"operator_id"`
	ModActionData ModActionData `json:"mod_action_data"`
	UserNoteData  UserNoteData  `json:"user_note_data"`
	Subreddit     string        `json:"subreddit"`
	User          string        `json:"user"`
	Operator      string        `json:"operator"`
	Id            string        `json:"id"`
	UserId        string        `json:"user_id"`
	CreatedAt     int           `json:"created_at"`
	Cursor        string        `json:"cursor"`
	Type          string        `json:"type"`
}

type notesList struct {
	Modnotes []*Modnote `json:"mod_notes"`
}

type ModnoteService struct {
	client *Client
}

type ModnoteFilterString string

const (
	ModnoteFilterStringNote          ModnoteFilterString = "NOTE"
	ModnoteFilterStringApproval      ModnoteFilterString = "APPROVAL"
	ModnoteFilterStringRemoval       ModnoteFilterString = "REMOVAL"
	ModnoteFilterStringBan           ModnoteFilterString = "BAN"
	ModnoteFilterStringMute          ModnoteFilterString = "MUTE"
	ModnoteFilterStringInvite        ModnoteFilterString = "INVITE"
	ModnoteFilterStringSpam          ModnoteFilterString = "SPAM"
	ModnoteFilterStringContentChange ModnoteFilterString = "CONTENT_CHANGE"
	ModnoteFilterStringModAction     ModnoteFilterString = "MOD_ACTION"
	ModnoteFilterStringAll           ModnoteFilterString = "ALL"
)

type ModnoteLabelString string

const (
	ModnoteLabelStringBan              ModnoteLabelString = "BAN"
	ModnoteLabelStringBotBan           ModnoteLabelString = "BOT_BAN"
	ModnoteLabelStringPermaBan         ModnoteLabelString = "PERMA_BAN"
	ModnoteLabelStringAbuseWarning     ModnoteLabelString = "ABUSE_WARNING"
	ModnoteLabelStringSpamWarning      ModnoteLabelString = "SPAM_WARNING"
	ModnoteLabelStringSpamWatch        ModnoteLabelString = "SPAM_WATCH"
	ModnoteLabelStringSolidContributor ModnoteLabelString = "SOLID_CONTRIBUTOR"
	ModnoteLabelStringHelpfulUser      ModnoteLabelString = "HELPFUL_USER"
)

type GetModnotesForUserOptions struct {
	// Before is an encoded pagination string. Notes have a "cursor" field that indicates what can go in this field
	Before *string              `url:"before,omitempty"`
	Filter *ModnoteFilterString `url:"filter,omitempty"`
	// Limit (default: 25, max: 100)
	Limit *int `url:"limit,omitempty"`
}

func (s *ModnoteService) GetModenotesForUser(ctx context.Context, subreddit string, user string, opts *GetModnotesForUserOptions) ([]*Modnote, *Response, error) {
	if opts == nil {
		opts = &GetModnotesForUserOptions{}
	}
	params := struct {
		Limit     *int                 `url:"limit,omitempty"`
		Filter    *ModnoteFilterString `url:"filter,omitempty"`
		Before    *string              `url:"before,omitempty"`
		User      string               `url:"user,omitempty"`
		Subreddit string               `url:"subreddit,omitempty"`
	}{
		Limit:     opts.Limit,
		Filter:    opts.Filter,
		Before:    opts.Before,
		User:      user,
		Subreddit: subreddit,
	}

	notes := &notesList{}
	resp, err := s.client.doReqWithOptions(ctx, http.MethodGet, "api/mod/notes", &params, notes)
	if err != nil {
		return nil, nil, err
	}

	return notes.Modnotes, resp, nil
}

type ModnoteUserSubredditPair struct {
	Subreddit string
	User      string
}

func (s *ModnoteService) GetRecentModenotesForPairs(ctx context.Context, pairs []*ModnoteUserSubredditPair) ([]*Modnote, *Response, error) {
	params := struct {
		Subreddits string `url:"subreddits,omitempty"`
		Users      string `url:"users,omitempty"`
	}{}
	var subreddits []string
	var users []string
	for _, pair := range pairs {
		subreddits = append(subreddits, pair.Subreddit)
		users = append(users, pair.User)
	}
	params.Subreddits = strings.Join(subreddits, ",")
	params.Users = strings.Join(users, ",")

	notes := &notesList{}
	resp, err := s.client.doReqWithOptions(ctx, http.MethodGet, "api/mod/notes/recent", &params, notes)
	if err != nil {
		return nil, nil, err
	}

	return notes.Modnotes, resp, nil
}

// DeleteModnote returns true if the note was deleted. Frankly I don't know when it would return false, since the API would return errors if it's not found or invalid
// If you know feel free to PR the change to this comment
func (s *ModnoteService) DeleteModnote(ctx context.Context, subreddit string, user string, noteID string) (bool, *Response, error) {
	params := struct {
		NoteID    string `url:"note_id,omitempty"`
		User      string `url:"user,omitempty"`
		Subreddit string `url:"subreddit,omitempty"`
	}{
		NoteID:    noteID,
		User:      user,
		Subreddit: subreddit,
	}

	deleted := &struct {
		Deleted bool `json:"deleted"`
	}{}
	resp, err := s.client.doReqWithOptions(ctx, http.MethodDelete, "api/mod/notes", &params, deleted)
	if err != nil {
		return false, nil, err
	}

	return deleted.Deleted, resp, nil
}

type CreateModnoteOptions struct {
	Label    *ModnoteLabelString
	RedditID *string
}

// CreateNote creates a new modnote. Specify a t3_.. or t1_.. id in the reddit_id field if you want to link to a specific post or comment
// Specify a label if you want to categorize the note.
func (s *ModnoteService) CreateModnote(ctx context.Context, subreddit string, user string, message string, opts *CreateModnoteOptions) (*Modnote, *Response, error) {
	if opts == nil {
		opts = &CreateModnoteOptions{}
	}
	params := struct {
		Label     *ModnoteLabelString `url:"label,omitempty"`
		User      string              `url:"user,omitempty"`
		Subreddit string              `url:"subreddit,omitempty"`
		Note      string              `url:"note,omitempty"`
		RedditID  *string             `url:"reddit_id,omitempty"`
	}{
		User:      user,
		Subreddit: subreddit,
		Note:      message,
		RedditID:  opts.RedditID,
		Label:     opts.Label,
	}

	created := &struct {
		Created *Modnote `json:"created"`
	}{}
	resp, err := s.client.doReqWithOptions(ctx, http.MethodPost, "api/mod/notes", &params, created)
	if err != nil {
		return nil, nil, err
	}

	return created.Created, resp, nil
}
