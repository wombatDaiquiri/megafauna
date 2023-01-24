package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mattn/go-mastodon"
	"gorm.io/gorm"
	"time"
)

func StatusFromMastodonEvent(status mastodon.Status) Status {
	mediaAttachments, mentions, tags := Attachments{}, []Mention{}, []Tag{}
	for _, attachment := range status.MediaAttachments {
		mediaAttachments = append(mediaAttachments, Attachment{
			Type:        attachment.Type,
			URL:         attachment.URL,
			RemoteURL:   attachment.RemoteURL,
			PreviewURL:  attachment.PreviewURL,
			TextURL:     attachment.TextURL,
			Description: attachment.Description,
		})
	}
	for _, mention := range status.Mentions {
		mentions = append(mentions, Mention{
			Model:    gorm.Model{},
			URL:      mention.URL,
			Username: mention.Username,
			Acct:     mention.Acct,
		})
	}
	for _, tag := range status.Tags {
		tags = append(tags, Tag{
			Model: gorm.Model{},
			Name:  tag.Name,
			URL:   tag.URL,
		})
	}

	return Status{
		Model:            gorm.Model{},
		URI:              status.URI,
		URL:              status.URL,
		Acct:             status.Account.Acct,
		Content:          status.Content,
		StatusCreatedAt:  status.CreatedAt,
		RepliesCount:     status.RepliesCount,
		ReblogsCount:     status.ReblogsCount,
		FavouritesCount:  status.FavouritesCount,
		Sensitive:        status.Sensitive,
		SpoilerText:      status.SpoilerText,
		Visibility:       status.Visibility,
		MediaAttachments: mediaAttachments,
		Mentions:         mentions,
		Tags:             tags,
		Language:         status.Language,
	}
}

// Status contains status data stored in megafauna.
type Status struct {
	gorm.Model
	URI  string
	URL  string
	Acct string

	Content         string
	StatusCreatedAt time.Time

	RepliesCount    int64
	ReblogsCount    int64
	FavouritesCount int64

	Sensitive        bool
	SpoilerText      string
	Visibility       string
	MediaAttachments Attachments
	Mentions         []Mention
	Tags             []Tag

	Language string
}

var (
	_ sql.Scanner   = (*Attachments)(nil)
	_ driver.Valuer = (*Attachments)(nil)
)

type Attachments []Attachment

// Scan scan value into Jsonb, implements sql.Scanner interface
func (attachments *Attachments) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return json.Unmarshal(bytes, attachments)
}

// Value return json value, implement driver.Valuer interface
func (attachments Attachments) Value() (driver.Value, error) {
	return json.Marshal(attachments)
}

// Attachment is stored as marshalled JSON.
type Attachment struct {
	Type        string
	URL         string
	RemoteURL   string
	PreviewURL  string
	TextURL     string
	Description string
}

type Mention struct {
	gorm.Model
	StatusID uint
	URL      string
	Username string
	Acct     string
}

type Tag struct {
	gorm.Model
	StatusID uint
	Name     string
	URL      string
}
