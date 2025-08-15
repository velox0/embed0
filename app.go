package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"golang.org/x/net/html"
)

var (
	minifiedJS []byte
)

type PageInfo struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	URL         string   `json:"url"`
	Theme       string   `json:"theme"`
}

var host string

// INIT //

func init() {
	// Initialize minifier and minify JS during startup //
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)

	// Read original JS file
	jsContent, err := os.ReadFile("static/embed.js")
	if err != nil {
		log.Fatal("Failed to read embed.js:", err)
	}

	jsWithHost := strings.ReplaceAll(string(jsContent), "{{HOST}}", os.Getenv("HOST"))

	minified, err := m.String("text/javascript", jsWithHost)
	if err != nil {
		log.Fatal("Failed to minify JS:", err)
	}

	minifiedJS = []byte(minified)
}

func main() {
	// ENV //
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using defaults")
	}

	host = os.Getenv("HOST")
	if host == "" {
		host = "http://localhost:8080"
		log.Println("No HOST in .env, defaulting to localhost")
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes //
	http.HandleFunc("/api/embed", withCORS(embedHandler))
	http.HandleFunc("/embed.js", withCORS(jsHandler))
	http.HandleFunc("/embed.css", withCORS(cssHandler))
	http.HandleFunc("/metagen", withCORS(metaGenHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s with host %s", port, host)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handlers //

func metaGenHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("url")
	if path == "" {
		http.Error(w, "Missing URL in path", http.StatusBadRequest)
		return
	}

	targetURL, err := url.PathUnescape(path)
	if err != nil {
		http.Error(w, "Invalid URL encoding", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	// Fetch metadata with enhanced content extraction
	title, description, images := fetchMetaEnhanced(targetURL)
	if len(images) == 0 {
		// Fallback image
		if u, err := url.Parse(targetURL); err == nil {
			images = append(images, "https://"+u.Host+"/favicon.ico")
		}
	}

	w.Header().Set("Content-Type", "text/html")
	html := generateMetaHTML(targetURL, title, description, images)
	w.Write([]byte(html))
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(minifiedJS) // Serve pre-minified version
}

func embedHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	theme := r.URL.Query().Get("theme")
	if url == "" {
		http.Error(w, "Missing ?url parameter", http.StatusBadRequest)
		return
	}
	if theme == "" {
		theme = os.Getenv("THEME")
		if theme == "" {
			theme = "light"
		}
	}

	title, desc, images := fetchMeta(url)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PageInfo{
		Title:       title,
		Description: desc,
		Images:      images,
		URL:         url,
		Theme:       theme,
	})
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	http.ServeFile(w, r, "static/embed.css")
}

// Utility functions //

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func fetchMeta(target string) (string, string, []string) {
	resp, err := http.Get(target)
	if err != nil {
		return target, "", nil
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return target, "", nil
	}

	var title, desc string
	var images []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "title" && n.FirstChild != nil && title == "" {
				title = strings.TrimSpace(n.FirstChild.Data)
			}
			if n.Data == "meta" {
				var name, property, content string
				for _, attr := range n.Attr {
					switch strings.ToLower(attr.Key) {
					case "name":
						name = strings.ToLower(attr.Val)
					case "property":
						property = strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					}
				}
				if (name == "description" || property == "og:description") && desc == "" {
					desc = content
				}
				if (property == "og:title" || name == "title") && title == "" {
					title = content
				}
				if (property == "og:image" || name == "twitter:image") && content != "" {
					images = append(images, content)
				}
			}

			if n.Data == "img" && len(images) == 0 {
				for _, attr := range n.Attr {
					if attr.Key == "src" && !strings.HasPrefix(attr.Val, "data:") {
						images = append(images, attr.Val)
						break // Stop after finding first image
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	if title == "" {
		if u, err := url.Parse(target); err == nil {
			title = u.Hostname()
		} else {
			title = "Untitled Page"
		}
	}

	if desc == "" {
		desc = "No description available"
	}

	// Fix image paths if needed
	if len(images) > 0 {
		baseURL, err := url.Parse(target)
		if err == nil {
			for i, img := range images {
				images[i] = fixURL(strings.TrimSpace(img), baseURL)
			}
		}
		images = uniqueStrings(images)
	}

	return title, desc, images
}

// fixURL converts relative or protocol-relative paths to absolute //
func fixURL(img string, base *url.URL) string {
	if img == "" {
		return ""
	}
	if strings.HasPrefix(img, "//") {
		return base.Scheme + ":" + img
	}
	if strings.HasPrefix(img, "/") {
		return base.Scheme + "://" + base.Host + img
	}
	if !strings.HasPrefix(img, "http://") && !strings.HasPrefix(img, "https://") {
		return base.Scheme + "://" + base.Host + "/" + img
	}
	return img
}

func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func fetchMetaEnhanced(target string) (string, string, []string) {
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with a browser-like User-Agent
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return target, "", nil
	}

	// Set a desktop Chrome User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return target, "", nil
	}
	defer resp.Body.Close()

	// Rest of your parsing logic...
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return target, "", nil
	}

	var (
		title         string
		description   string
		images        []string
		paragraphs    []string
		foundFirstImg bool
	)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && title == "" {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "meta":
				processMetaTag(n, &title, &description, &images)
			case "h1", "h2", "h3":
				if text := extractText(n); text != "" && description == "" {
					paragraphs = append(paragraphs, text)
				}
			case "p":
				if text := extractText(n); text != "" && len(paragraphs) < 3 { // Limit # of paragraphs
					paragraphs = append(paragraphs, text)
				}
			case "img":
				if len(images) == 0 && !foundFirstImg {
					if src := getImgSrc(n); src != "" {
						images = append(images, src)
						foundFirstImg = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// Fallback title from URL
	if title == "" {
		if u, err := url.Parse(target); err == nil {
			title = u.Hostname()
		} else {
			title = "Untitled Page"
		}
	}

	// Build description from content if no meta description
	if description == "" && len(paragraphs) > 0 {
		description = strings.Join(paragraphs, " ")
		if len(description) > 200 { // Truncate long descriptions
			description = description[:200] + "..."
		}
	} else if description == "" {
		description = "No description available"
	}

	// Fix image URLs
	if len(images) > 0 {
		baseURL, err := url.Parse(target)
		if err == nil {
			for i, img := range images {
				images[i] = fixURL(strings.TrimSpace(img), baseURL)
			}
		}
	}

	return title, description, images
}

func processMetaTag(n *html.Node, title, description *string, images *[]string) {
	var name, property, content string
	for _, attr := range n.Attr {
		switch strings.ToLower(attr.Key) {
		case "name":
			name = strings.ToLower(attr.Val)
		case "property":
			property = strings.ToLower(attr.Val)
		case "content":
			content = attr.Val
		}
	}

	if content == "" {
		return
	}

	if (name == "description" || property == "og:description") && *description == "" {
		*description = content
	}
	if (property == "og:title" || name == "title") && *title == "" {
		*title = content
	}
	if property == "og:image" || name == "twitter:image" {
		*images = append(*images, content)
	}
}

func extractText(n *html.Node) string {
	var buf bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			buf.WriteString(c.Data)
		case html.ElementNode:
			buf.WriteString(extractText(c))
		}
	}
	return strings.TrimSpace(buf.String())
}

func getImgSrc(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "src" && !strings.HasPrefix(attr.Val, "data:") {
			return attr.Val
		}
	}
	return ""
}

func generateMetaHTML(url, title, description string, images []string) string {
	imageURL := ""
	if len(images) > 0 {
		imageURL = images[0]
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>%s</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:type" content="website">
    <meta property="og:url" content="%s">
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:image" content="%s">
    
    <!-- Twitter -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">
</head>
<body>
    <h1>%s</h1>
    <p>%s</p>
    <a href="%s">View original page</a>
</body>
</html>`,
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(description),
		template.HTMLEscapeString(url),
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(description),
		template.HTMLEscapeString(imageURL),
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(description),
		template.HTMLEscapeString(imageURL),
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(description),
		template.HTMLEscapeString(url))
}
