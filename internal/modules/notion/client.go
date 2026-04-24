package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	notionVersion = "2022-06-28"
	notionBase    = "https://api.notion.com/v1"
)

// notionClient performs authenticated calls to the Notion API.
type notionClient struct {
	token string
	http  *http.Client
}

func newClient(token string) *notionClient {
	return &notionClient{token: token, http: &http.Client{}}
}

// richText is a subset of Notion's rich text object.
type richText struct {
	PlainText string `json:"plain_text"`
}

// statusValue is the inner object for Notion's "status" property type.
type statusValue struct {
	Name string `json:"name"`
}

// selectValue is the inner object for Notion's "select" property type.
type selectValue struct {
	Name string `json:"name"`
}

// personValue is a single user inside a "people" property.
type personValue struct {
	Person struct {
		Email string `json:"email"`
	} `json:"person"`
}

// notionFile represents one item inside a "files" property.
type notionFile struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"` // "file" or "external"
	File     notionFileInner `json:"file"`
	External notionFileInner `json:"external"`
}

type notionFileInner struct {
	URL string `json:"url"`
}

// PageFile is the parsed file info returned to callers.
type PageFile struct {
	Name string
	URL  string
}

// property represents a Notion page property value.
type property struct {
	Type     string        `json:"type"`
	Title    []richText    `json:"title"`
	RichText []richText    `json:"rich_text"`
	Status   statusValue   `json:"status"`
	Select   selectValue   `json:"select"`
	Checkbox bool          `json:"checkbox"`
	Files    []notionFile  `json:"files"`
	People   []personValue `json:"people"`
}

// page is a single Notion database row.
type page struct {
	ID         string              `json:"id"`
	Properties map[string]property `json:"properties"`
}

// queryResponse is the body returned by POST /databases/{id}/query.
type queryResponse struct {
	Results    []page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// dbStatusOption is one option inside a Notion database "status" or "select" property schema.
type dbStatusOption struct {
	Name string `json:"name"`
}

// dbStatusSchema holds the options list for a "status"-type database property.
type dbStatusSchema struct {
	Options []dbStatusOption `json:"options"`
}

// dbSelectSchema holds the options list for a "select"-type database property.
type dbSelectSchema struct {
	Options []dbStatusOption `json:"options"`
}

// dbProperty is an entry inside the database-level "properties" map.
type dbProperty struct {
	Type   string         `json:"type"`
	Status dbStatusSchema `json:"status"`
	Select dbSelectSchema `json:"select"`
}

// databaseResponse is the body returned by GET /databases/{id}.
type databaseResponse struct {
	Title      []richText            `json:"title"`
	Properties map[string]dbProperty `json:"properties"`
}

func (c *notionClient) do(method, url string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("notion responded %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// getDatabase returns the title and all status/select option labels for a Notion database.
func (c *notionClient) getDatabase(databaseID string) (title string, statusLabels []string, err error) {
	data, err := c.do("GET", notionBase+"/databases/"+databaseID, nil)
	if err != nil {
		return "", nil, err
	}
	var db databaseResponse
	if err := json.Unmarshal(data, &db); err != nil {
		return "", nil, err
	}
	title = joinRichText(db.Title)

	// Collect all status/select options from the database schema.
	// Priority: properties named "Status" first, then any other status/select props.
	seen := map[string]bool{}
	addOption := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			statusLabels = append(statusLabels, name)
		}
	}

	// Pass 1: property keyed "Status" (any type — status or select)
	for key, prop := range db.Properties {
		if !strings.EqualFold(key, "status") {
			continue
		}
		switch prop.Type {
		case "status":
			for _, o := range prop.Status.Options {
				addOption(o.Name)
			}
		case "select":
			for _, o := range prop.Select.Options {
				addOption(o.Name)
			}
		}
	}
	// Pass 2: any property of Notion's native "status" type (not arbitrary select).
	// This covers databases where the status column has a non-"Status" name.
	// We intentionally skip "select" here — select can be Priority, Category, etc.
	for key, prop := range db.Properties {
		if strings.EqualFold(key, "status") {
			continue // already handled in pass 1
		}
		if prop.Type == "status" {
			for _, o := range prop.Status.Options {
				addOption(o.Name)
			}
		}
	}

	return title, statusLabels, nil
}

// getDatabaseTitle returns the plain-text title of a Notion database.
// Kept for compatibility; prefer getDatabase when status labels are also needed.
func (c *notionClient) getDatabaseTitle(databaseID string) (string, error) {
	title, _, err := c.getDatabase(databaseID)
	return title, err
}

// queryDatabase fetches all pages from a Notion database (handles pagination).
func (c *notionClient) queryDatabase(databaseID string) ([]page, error) {
	var pages []page
	cursor := ""

	for {
		payload := map[string]any{"page_size": 100}
		if cursor != "" {
			payload["start_cursor"] = cursor
		}

		data, err := c.do("POST", notionBase+"/databases/"+databaseID+"/query", payload)
		if err != nil {
			return nil, err
		}

		var qr queryResponse
		if err := json.Unmarshal(data, &qr); err != nil {
			return nil, err
		}
		pages = append(pages, qr.Results...)

		if !qr.HasMore {
			break
		}
		cursor = qr.NextCursor
	}
	return pages, nil
}

// blockRichText is the rich_text array shared by many block types.
type blockRichText struct {
	PlainText string `json:"plain_text"`
}

// block represents one Notion block from GET /blocks/{id}/children.
type block struct {
	Type      string `json:"type"`
	Paragraph struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"paragraph"`
	Heading1 struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"heading_1"`
	Heading2 struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"heading_2"`
	Heading3 struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"heading_3"`
	BulletedListItem struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"bulleted_list_item"`
	NumberedListItem struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"numbered_list_item"`
	Quote struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"quote"`
	Callout struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"callout"`
	ToDo struct {
		RichText []blockRichText `json:"rich_text"`
	} `json:"to_do"`
}

type blocksResponse struct {
	Results []block `json:"results"`
}

// getPageText fetches the block children of a page and returns their plain text
// joined by newlines, suitable for use as a task description.
func (c *notionClient) getPageText(pageID string) (string, error) {
	data, err := c.do("GET", notionBase+"/blocks/"+pageID+"/children?page_size=100", nil)
	if err != nil {
		return "", err
	}
	var br blocksResponse
	if err := json.Unmarshal(data, &br); err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, b := range br.Results {
		var parts []blockRichText
		switch b.Type {
		case "paragraph":
			parts = b.Paragraph.RichText
		case "heading_1":
			parts = b.Heading1.RichText
		case "heading_2":
			parts = b.Heading2.RichText
		case "heading_3":
			parts = b.Heading3.RichText
		case "bulleted_list_item":
			parts = b.BulletedListItem.RichText
		case "numbered_list_item":
			parts = b.NumberedListItem.RichText
		case "quote":
			parts = b.Quote.RichText
		case "callout":
			parts = b.Callout.RichText
		case "to_do":
			parts = b.ToDo.RichText
		}
		line := ""
		for _, rt := range parts {
			line += rt.PlainText
		}
		if line != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(line)
		}
	}
	return sb.String(), nil
}

// joinRichText concatenates plain text segments into a single string.
func joinRichText(parts []richText) string {
	var sb strings.Builder
	for _, p := range parts {
		sb.WriteString(p.PlainText)
	}
	return strings.TrimSpace(sb.String())
}

// extractTitle finds the first title-type property and returns its plain text.
func extractTitle(p page) string {
	for _, prop := range p.Properties {
		if prop.Type == "title" {
			return joinRichText(prop.Title)
		}
	}
	return ""
}

// extractDescription finds the first rich_text property named "Description" (case-insensitive).
func extractDescription(p page) string {
	for key, prop := range p.Properties {
		if prop.Type == "rich_text" && strings.EqualFold(key, "description") {
			return joinRichText(prop.RichText)
		}
	}
	return ""
}

// extractRawStatus returns the raw status label from the Notion page.
// Priority:
//  1. property whose KEY is exactly "Status" (case-insensitive, any type)
//  2. any other property of Notion's native "status" type
//  3. checkbox as last resort
//
// Arbitrary "select" properties (e.g. Priority, Category) are intentionally
// ignored to avoid treating non-status fields as task statuses.
func extractRawStatus(p page) string {
	// pass 1: property named "Status"
	for key, prop := range p.Properties {
		if !strings.EqualFold(key, "status") {
			continue
		}
		switch prop.Type {
		case "status":
			if prop.Status.Name != "" {
				return prop.Status.Name
			}
		case "select":
			if prop.Select.Name != "" {
				return prop.Select.Name
			}
		}
	}
	// pass 2: any other property of Notion's native "status" type
	for key, prop := range p.Properties {
		if strings.EqualFold(key, "status") {
			continue
		}
		if prop.Type == "status" && prop.Status.Name != "" {
			return prop.Status.Name
		}
	}
	// pass 3: checkbox
	for _, prop := range p.Properties {
		if prop.Type == "checkbox" {
			if prop.Checkbox {
				return "Done"
			}
			return "Not started"
		}
	}
	return ""
}

// propTypes returns a summary of property key→type pairs for debug logging.
func propTypes(p page) string {
	var b strings.Builder
	for key, prop := range p.Properties {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(prop.Type)
		// include status/select name if present
		switch prop.Type {
		case "status":
			b.WriteString("(")
			b.WriteString(prop.Status.Name)
			b.WriteString(")")
		case "select":
			b.WriteString("(")
			b.WriteString(prop.Select.Name)
			b.WriteString(")")
		case "checkbox":
			if prop.Checkbox {
				b.WriteString("(true)")
			} else {
				b.WriteString("(false)")
			}
		}
	}
	return b.String()
}

// extractAssigneeEmail returns the email of the first person in the "Assignee"
// (or any "people"-type) property. Returns "" if no assignee is set.
func extractAssigneeEmail(p page) string {
	// Pass 1: property keyed "Assignee" (case-insensitive)
	for key, prop := range p.Properties {
		if prop.Type != "people" || !strings.EqualFold(key, "assignee") {
			continue
		}
		if len(prop.People) > 0 {
			return prop.People[0].Person.Email
		}
	}
	// Pass 2: any other people property
	for key, prop := range p.Properties {
		if prop.Type != "people" || strings.EqualFold(key, "assignee") {
			continue
		}
		if len(prop.People) > 0 {
			return prop.People[0].Person.Email
		}
	}
	return ""
}

// extractFiles collects all file URLs from "files"-type properties of a page.
func extractFiles(p page) []PageFile {
	var result []PageFile
	for _, prop := range p.Properties {
		if prop.Type != "files" {
			continue
		}
		for _, f := range prop.Files {
			var url string
			switch f.Type {
			case "file":
				url = f.File.URL
			case "external":
				url = f.External.URL
			}
			if url != "" {
				result = append(result, PageFile{Name: f.Name, URL: url})
			}
		}
	}
	return result
}

// codeFromName converts a human-readable status name into a slug code.
// "In progress" → "in_progress", "Not started" → "not_started".
func codeFromName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-':
			b.WriteRune('_')
		}
	}
	return b.String()
}
