package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s with host %s", port, host)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handlers //

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
