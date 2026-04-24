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
	Type     string       `json:"type"`
	Title    []richText   `json:"title"`
	RichText []richText   `json:"rich_text"`
	Status   statusValue  `json:"status"`
	Select   selectValue  `json:"select"`
	Checkbox bool         `json:"checkbox"`
	Files    []notionFile `json:"files"`
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

// databaseResponse is the body returned by GET /databases/{id}.
type databaseResponse struct {
	Title []richText `json:"title"`
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

// getDatabaseTitle returns the plain-text title of a Notion database.
func (c *notionClient) getDatabaseTitle(databaseID string) (string, error) {
	data, err := c.do("GET", notionBase+"/databases/"+databaseID, nil)
	if err != nil {
		return "", err
	}
	var db databaseResponse
	if err := json.Unmarshal(data, &db); err != nil {
		return "", err
	}
	return joinRichText(db.Title), nil
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
//  1. property whose KEY is exactly "Status" (case-insensitive)
//  2. any other status-type property
//  3. any select-type property
//  4. checkbox as last resort
func extractRawStatus(p page) string {
	// pass 1: property named "Status" — most reliable, avoids "Status 1" etc.
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
	// pass 2: any other status-type property
	for _, prop := range p.Properties {
		if prop.Type == "status" && prop.Status.Name != "" {
			return prop.Status.Name
		}
	}
	// pass 3: any select property
	for _, prop := range p.Properties {
		if prop.Type == "select" && prop.Select.Name != "" {
			return prop.Select.Name
		}
	}
	// pass 4: checkbox
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
