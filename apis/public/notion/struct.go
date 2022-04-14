package notion

import (
	"encoding/json"

	"golang.org/x/time/rate"
)

type webClient struct {
	token   string
	limiter *rate.Limiter
}

type PageDataRequest struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Blockid    string `json:"blockId"`
	Showmoveto bool   `json:"showMoveTo"`
	Saveparent bool   `json:"saveParent"`
}

type PageDataResponse struct {
	Pageid                string
	Cursor                *Block
	Spacename             string `json:"spaceName"`
	Spaceid               string `json:"spaceId"`
	Canjoinspace          bool   `json:"canJoinSpace"`
	Icon                  string `json:"icon"`
	Userhasexplicitaccess bool   `json:"userHasExplicitAccess"`
	Haspublicaccess       bool   `json:"hasPublicAccess"`
	Owneruserid           string `json:"ownerUserId"`
	Betaenabled           bool   `json:"betaEnabled"`
	Canrequestaccess      bool   `json:"canRequestAccess"`
}

// /api/v3/loadPageChunk request
type loadPageChunkRequest struct {
	PageID          string `json:"pageId"`
	ChunkNumber     int    `json:"chunkNumber"`
	Limit           int    `json:"limit"`
	Cursor          cursor `json:"cursor"`
	VerticalColumns bool   `json:"verticalColumns"`
}

type cursor struct {
	Stack [][]stack `json:"stack"`
}

type stack struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Table string `json:"table"`
}

// LoadPageChunkResponse is a response to /api/v3/loadPageChunk api
type LoadPageChunkResponse struct {
	RecordMap *RecordMap `json:"recordMap"`
	Cursor    cursor     `json:"cursor"`

	RawJSON map[string]any `json:"-"`
}

type submitTransactionRequest struct {
	RequestID   string        `json:"requestId"`
	Transaction []Transaction `json:"transactions"`
}

type Transaction struct {
	ID         string       `json:"id"`
	SpaceID    string       `json:"spaceId"`
	Operations []*Operation `json:"operations"`
}

type Pointer struct {
	ID    string `json:"id"`
	Table string `json:"table"`
}

type Operation struct {
	Point   Pointer  `json:"pointer"`
	Path    []string `json:"path"`
	Command string   `json:"command"`
	Args    any      `json:"args"`
}

// RecordMap contains a collections of blocks, a space, users, and collections.
type RecordMap struct {
	Activities      map[string]*Record `json:"activity"`
	Blocks          map[string]*Record `json:"block"`
	Spaces          map[string]*Record `json:"space"`
	Users           map[string]*Record `json:"notion_user"`
	Collections     map[string]*Record `json:"collection"`
	CollectionViews map[string]*Record `json:"collection_view"`
	Comments        map[string]*Record `json:"comment"`
	Discussions     map[string]*Record `json:"discussion"`
}

// POST /api/v3/getUploadFileUrl request
type getUploadFileUrlRequest struct {
	Bucket      string `json:"bucket"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

// GetUploadFileUrlResponse is a response to POST /api/v3/getUploadFileUrl
type GetUploadFileUrlResponse struct {
	URL          string         `json:"url"`
	SignedGetURL string         `json:"signedGetUrl"`
	SignedPutURL string         `json:"signedPutUrl"`
	FileID       string         `json:"-"`
	RawJSON      map[string]any `json:"-"`
}

// Record represents a polymorphic record
type Record struct {
	// fields returned by the server
	Role string `json:"role"`
	// polymorphic value of the record, which we decode into Block, Space etc.
	Value json.RawMessage `json:"value"`

	// fields set from Value based on type
	ID    string `json:"-"`
	Table string `json:"-"`
	Block *Block `json:"-"`
	// TODO: add more types
}

// Block describes a block
type Block struct {
	// values that come from JSON
	// a unique ID of the block
	ID string `json:"id"`
	// if false, the page is deleted
	Alive bool `json:"alive"`
	// List of block ids for that make up content of this block
	// Use Content to get corresponding block (they are in the same order)
	ContentIDs   []string `json:"content,omitempty"`
	CopiedFrom   string   `json:"copied_from,omitempty"`
	CollectionID string   `json:"collection_id,omitempty"` // for BlockCollectionView
	// ID of the user who created this block
	CreatedBy   string `json:"created_by"`
	CreatedTime int64  `json:"created_time"`

	CreatedByTable    string `json:"created_by_table"`     // e.g. "notion_user"
	CreatedByID       string `json:"created_by_id"`        // e.g. "bb760e2d-d679-4b64-b2a9-03005b21870a",
	LastEditedByTable string `json:"last_edited_by_table"` // e.g. "notion_user"
	LastEditedByID    string `json:"last_edited_by_id"`    // e.g. "bb760e2d-d679-4b64-b2a9-03005b21870a"

	// List of block ids with discussion content
	DiscussionIDs []string `json:"discussion,omitempty"`
	// those ids seem to map to storage in s3
	// https://s3-us-west-2.amazonaws.com/secure.notion-static.com/${id}/${name}
	FileIDs []string `json:"file_ids,omitempty"`

	// TODO: don't know what this means
	IgnoreBlockCount bool `json:"ignore_block_count,omitempty"`

	// ID of the user who last edited this block
	LastEditedBy   string `json:"last_edited_by"`
	LastEditedTime int64  `json:"last_edited_time"`
	// ID of parent Block
	ParentID    string         `json:"parent_id"`
	ParentTable string         `json:"parent_table"`
	Properties  map[string]any `json:"properties,omitempty"`
	// type of the block e.g. TypeText, TypePage etc.
	Type string `json:"type"`
	// blocks are versioned
	Version int64 `json:"version"`
	// for BlockCollectionView
	ViewIDs []string `json:"view_ids,omitempty"`

	// Parent of this block
	Parent *Block `json:"-"`

	// maps ContentIDs array to Block type
	Content []*Block `json:"-"`

	// for BlockPage
	Title string `json:"-"`

	// For BlockTodo, a checked state
	IsChecked bool `json:"-"`

	// for BlockBookmark
	Description string `json:"-"`
	Link        string `json:"-"`

	// for BlockBookmark it's the url of the page
	// for BlockGist it's the url for the gist
	// fot BlockImage it's url of the image, but use ImageURL instead
	// because Source is sometimes not accessible
	// for BlockFile it's url of the file
	// for BlockEmbed it's url of the embed
	Source string `json:"-"`

	// for BlockFile
	FileSize string `json:"-"`

	// for BlockImage it's an URL built from Source that is always accessible
	ImageURL string `json:"-"`

	// for BlockCode
	Code         string `json:"-"`
	CodeLanguage string `json:"-"`

	// RawJSON represents Block as
	RawJSON map[string]any `json:"-"`

	isResolved bool
}
