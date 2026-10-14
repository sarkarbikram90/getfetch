package normalizer

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

var (
	// Multi-whitespace matcher
	whitespaceRegex = regexp.MustCompile(`[ \t\r\f\v]+`)
	newlineRegex    = regexp.MustCompile(`\n{3,}`)

	// Tracking query parameters to strip
	trackingParams = map[string]bool{
		"utm_source":   true,
		"utm_medium":   true,
		"utm_campaign": true,
		"utm_term":     true,
		"utm_content":  true,
		"gh_src":       true,
		"lever-source": true,
		"ashby_jid":    true,
		"ref":          true,
		"source":       true,
		"fbclid":       true,
		"gclid":        true,
	}

	// Block-level HTML elements that should emit a newline or space
	blockTags = map[string]bool{
		"p":      true,
		"div":    true,
		"br":     true,
		"li":     true,
		"h1":     true,
		"h2":     true,
		"h3":     true,
		"h4":     true,
		"h5":     true,
		"h6":     true,
		"tr":     true,
		"hr":     true,
		"header": true,
		"footer": true,
		"block":  true,
	}
)

// StripHTML converts HTML content into clean, readable plain text.
// It removes style/script tags, preserves spacing across block elements,
// and normalizes multiple spaces/newlines.
func StripHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fallback to simple regex-based replacement if html parsing fails
		re := regexp.MustCompile(`<[^>]*>`)
		clean := re.ReplaceAllString(htmlContent, " ")
		return cleanWhitespace(html.UnescapeString(clean))
	}

	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Skip script and style tags completely
			if n.Data == "script" || n.Data == "style" || n.Data == "noscript" {
				return
			}
		}

		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}

		// Insert newline after block tags for readable separation
		if n.Type == html.ElementNode && blockTags[n.Data] {
			sb.WriteString("\n")
		}
	}

	walk(doc)
	return cleanWhitespace(html.UnescapeString(sb.String()))
}

// cleanWhitespace cleans up irregular spaces, tabs, and duplicate newlines.
func cleanWhitespace(text string) string {
	// Standardize spaces per line
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(whitespaceRegex.ReplaceAllString(line, " "))
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	result := strings.Join(cleanedLines, "\n")
	return newlineRegex.ReplaceAllString(result, "\n\n")
}

// NormalizeURL strips tracking parameters, lowercases host, and ensures consistent format.
func NormalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Lowercase scheme and host
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	// Clean path trailing slash (unless root path)
	if len(parsed.Path) > 1 && strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	// Filter out tracking query parameters
	values := parsed.Query()
	for param := range values {
		if trackingParams[strings.ToLower(param)] {
			values.Del(param)
		}
	}

	// Sort query parameters for stable canonical representation
	if len(values) > 0 {
		var keys []string
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var queryParts []string
		for _, k := range keys {
			for _, v := range values[k] {
				queryParts = append(queryParts, url.QueryEscape(k)+"="+url.QueryEscape(v))
			}
		}
		parsed.RawQuery = strings.Join(queryParts, "&")
	} else {
		parsed.RawQuery = ""
	}

	// Strip fragment/hash
	parsed.Fragment = ""

	return parsed.String()
}

// ComputeContentHash generates a SHA256 signature for deduplication and change detection.
func ComputeContentHash(title, company, descriptionText string) string {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(strings.TrimSpace(title))))
	h.Write([]byte("|"))
	h.Write([]byte(strings.ToLower(strings.TrimSpace(company))))
	h.Write([]byte("|"))
	h.Write([]byte(strings.TrimSpace(descriptionText)))
	return hex.EncodeToString(h.Sum(nil))
}
