package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"golang.org/x/net/html"
)

const version = "v1.2.0"

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
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)

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
	// ENV
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

	http.HandleFunc("/", withCORS(homeHandler))
	http.HandleFunc("/api/embed", withCORS(embedHandler))
	http.HandleFunc("/embed.js", withCORS(jsHandler))
	http.HandleFunc("/embed.css", withCORS(cssHandler))
	http.HandleFunc("/api/metagen", withCORS(metaGenHandler))
	http.HandleFunc("/version", withCORS(versionHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s with host %s", port, host)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handlers

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

	ua := r.URL.Query().Get("ua")

	// Use enhanced fetch with headless fallback
	title, description, images := fetchMetaEnhanced(targetURL, ua)
	if len(images) == 0 {
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
	w.Write(minifiedJS)
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

	ua := r.URL.Query().Get("ua")

	// Use enhanced fetch with headless fallback
	title, desc, images := fetchMetaEnhanced(url, ua)
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

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(version))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(homePageHTML))
}

const homePageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>embed0 :: metadata extraction server</title>
<style>
  *{margin:0;padding:0;box-sizing:border-box}
  body{
    background:#0a0a0a;color:#00ff41;
    font-family:"SF Mono","Fira Code","Cascadia Code",monospace;
    font-size:14px;line-height:1.6;
    padding:20px;max-width:900px;margin:0 auto;
  }
  a{color:#00ccff;text-decoration:none}
  a:hover{text-decoration:underline}

  .header{
    border:1px dashed #00ff41;padding:20px;margin-bottom:24px;
    text-align:center;
  }
  .header h1{font-size:20px;font-weight:400;margin-bottom:8px}
  .header .version{color:#666;font-size:12px}

  .packet{
    border:1px solid #333;padding:16px;margin-bottom:16px;
    position:relative;
  }
  .packet::before{
    content:"[PACKET]";position:absolute;top:-9px;left:12px;
    background:#0a0a0a;padding:0 6px;font-size:10px;color:#666;
  }

  .divider{
    border:none;border-top:1px dashed #333;
    margin:20px 0;
  }

  .section-title{
    color:#ffcc00;font-size:12px;text-transform:uppercase;
    letter-spacing:2px;margin-bottom:12px;
  }

  .endpoint{
    display:flex;align-items:center;gap:12px;
    padding:8px 0;border-bottom:1px solid #1a1a1a;
  }
  .endpoint:last-child{border-bottom:none}
  .method{
    background:#00ff41;color:#000;padding:2px 8px;
    font-size:11px;font-weight:700;min-width:50px;text-align:center;
  }
  .method.get{background:#00ff41}
  .method.post{background:#ffcc00}
  .path{color:#00ccff;min-width:140px}
  .desc{color:#888;font-size:12px}

  .explorer{margin-top:8px}
  .explorer label{color:#666;font-size:11px;display:block;margin-bottom:4px}
  .explorer input,.explorer select{
    background:#111;border:1px solid #333;color:#00ff41;
    font-family:inherit;font-size:13px;
    padding:8px 12px;width:100%;margin-bottom:12px;
  }
  .explorer input:focus,.explorer select:focus{
    outline:none;border-color:#00ff41;
  }
  .explorer button{
    background:#00ff41;color:#000;border:none;
    font-family:inherit;font-size:13px;font-weight:700;
    padding:8px 24px;cursor:pointer;text-transform:uppercase;
  }
  .explorer button:hover{background:#00cc33}
  .explorer button:active{transform:scale(0.98)}

  .result{
    background:#111;border:1px solid #333;
    padding:16px;margin-top:16px;
    font-size:12px;max-height:400px;overflow-y:auto;
    display:none;
  }
  .result.visible{display:block}
  .result .label{color:#666;font-size:10px;text-transform:uppercase;letter-spacing:1px}
  .result .status{color:#00ff41;padding-bottom:8px;border-bottom:1px dashed #333}
  .result .status.err{color:#ff4444}
  .result .body{color:#ccc;white-space:pre-wrap;word-break:break-all;padding-top:8px}

  .footer{
    border:1px dashed #333;padding:16px;margin-top:24px;
    text-align:center;color:#555;font-size:11px;
  }

  .ascii-art{
    color:#333;font-size:11px;line-height:1.2;
    text-align:center;margin-bottom:16px;
  }

  .stats{
    display:grid;grid-template-columns:repeat(3,1fr);gap:12px;
    margin-bottom:16px;
  }
  .stat{
    border:1px solid #222;padding:12px;text-align:center;
  }
  .stat .num{color:#00ff41;font-size:18px}
  .stat .label{color:#555;font-size:10px;text-transform:uppercase;margin-top:4px}

  @media(max-width:600px){
    .stats{grid-template-columns:1fr}
    .endpoint{flex-direction:column;align-items:flex-start;gap:4px}
  }
  .img-panel details{margin-top:12px}
  .img-panel summary{color:#ffcc00;cursor:pointer;font-size:12px}
  .img-panel img{max-width:400px;max-height:250px;margin:8px 0;display:block;border:1px solid #333}
</style>
</head>
<body>

<div class="header">
  <pre class="ascii-art">
   ___           ___           ___           ___           ___
   /\  \         /\__\         /\  \         /\  \         /\  \
   /::\  \       /::|  |       /::\  \       /::\  \       /::\  \
   /:/\:\  \     /:|:|  |      /:/\:\  \     /:/\:\  \     /:/\:\  \
   /::\~\:\  \   /:/|:|__|__   /::\~\:\__\   /::\~\:\  \   /:/  \:\__\
   /:/\:\ \:\__\ /:/ |::::\__\ /:/\:\ \:|__| /:/\:\ \:\__\ /:/__/ \:|__|
   \:\~\:\ \/__/ \/__/~~/:/  / \:\~\:\/:/  / \:\~\:\ \/__/ \:\  \ /:/  /
   \:\ \:\__\         /:/  /   \:\ \::/  /   \:\ \:\__\    \:\  /:/  /
   \:\ \/__/        /:/  /     \:\/:/  /     \:\ \/__/     \:\/:/  /
   \:\__\         /:/  /       \::/__/       \:\__\        \::/__/
\/__/         \/__/         ~~            \/__/         ~~
  </pre>
  <h1>embed0</h1>
  <div class="version">` + version + ` // metadata extraction server</div>
</div>

<div class="stats">
  <div class="stat">
    <div class="num">5</div>
    <div class="label">Endpoints</div>
  </div>
  <div class="stat">
    <div class="num">3</div>
    <div class="label">Themes</div>
  </div>
  <div class="stat">
    <div class="num">5</div>
    <div class="label">UA Presets</div>
  </div>
</div>

<div class="packet">
  <div class="section-title">// endpoints</div>
  <div class="endpoint">
    <span class="method get">GET</span>
    <span class="path">/</span>
    <span class="desc">API explorer (this page)</span>
  </div>
  <div class="endpoint">
    <span class="method get">GET</span>
    <span class="path">/api/embed</span>
    <span class="desc">JSON metadata for target URL</span>
  </div>
  <div class="endpoint">
    <span class="method get">GET</span>
    <span class="path">/api/metagen</span>
    <span class="desc">HTML page with OG/Twitter meta tags</span>
  </div>
  <div class="endpoint">
    <span class="method get">GET</span>
    <span class="path">/embed.js</span>
    <span class="desc">Embeddable widget script</span>
  </div>
  <div class="endpoint">
    <span class="method get">GET</span>
    <span class="path">/embed.css</span>
    <span class="desc">Widget stylesheet</span>
  </div>
  <div class="endpoint">
    <span class="method get">GET</span>
    <span class="path">/version</span>
    <span class="desc">Server version string</span>
  </div>
</div>

<hr class="divider">

<div class="packet">
  <div class="section-title">// api explorer</div>
  <div class="explorer">
    <label>TARGET URL</label>
    <input type="text" id="targetUrl" placeholder="https://github.com" value="https://github.com">

    <label>ENDPOINT</label>
    <select id="endpoint">
      <option value="api">/api/embed (JSON)</option>
      <option value="metagen">/api/metagen (HTML)</option>
      <option value="version">/version (text)</option>
    </select>

    <label>THEME</label>
    <select id="theme">
      <option value="light">light</option>
      <option value="dark">dark</option>
      <option value="minimal">minimal</option>
    </select>

    <label>USER-AGENT</label>
    <select id="ua">
      <option value="">default</option>
      <option value="mobile">mobile</option>
      <option value="bot">bot (Googlebot)</option>
      <option value="twitter">twitter</option>
      <option value="facebook">facebook</option>
      <option value="discord">discord</option>
    </select>

    <button onclick="sendRequest()">TRANSMIT</button>

    <div class="result" id="result">
      <div class="label">response</div>
      <div class="status" id="status"></div>
      <div class="body" id="body"></div>
    </div>
  </div>
</div>

<hr class="divider">

<div class="packet">
  <div class="section-title">// quick usage</div>
  <div style="color:#888;font-size:12px;line-height:1.8">
    <div><span style="color:#00ccff">curl</span> "http://HOST/api/embed?url=https://example.com"</div>
    <div><span style="color:#00ccff">curl</span> "http://HOST/api/embed?url=https://example.com&theme=dark&ua=twitter"</div>
    <div><span style="color:#00ccff">curl</span> "http://HOST/api/metagen?url=https://example.com"</div>
    <div><span style="color:#00ccff">curl</span> "http://HOST/version"</div>
  </div>
</div>

<hr class="divider">

<div class="packet">
  <div class="section-title">// embed widget</div>
  <div style="color:#888;font-size:12px;line-height:1.8">
    <div style="color:#666">// add to any HTML page:</div>
    <div>&lt;<span style="color:#ffcc00">link</span> rel="stylesheet" href="http://HOST/embed.css"&gt;</div>
    <div>&lt;<span style="color:#ffcc00">div</span> data-url="https://example.com" data-theme="dark"&gt;&lt;/div&gt;</div>
    <div>&lt;<span style="color:#ffcc00">script</span> async src="http://HOST/embed.js"&gt;&lt;/script&gt;</div>
  </div>
</div>

<div class="footer">
  embed0 v` + version + ` // net/http // Go // chromedp<br>
  <a href="https://github.com/velox0/embed0">github.com/velox0/embed0</a>
</div>

<script>
function sendRequest(){
  var url=document.getElementById("targetUrl").value;
  var ep=document.getElementById("endpoint").value;
  var theme=document.getElementById("theme").value;
  var ua=document.getElementById("ua").value;
  var result=document.getElementById("result");
  var status=document.getElementById("status");
  var body=document.getElementById("body");

  if(!url){alert("Enter a URL");return}

  var reqUrl="";
  if(ep==="api"){
    reqUrl="/api/embed?url="+encodeURIComponent(url)+"&theme="+encodeURIComponent(theme);
    if(ua)reqUrl+="&ua="+encodeURIComponent(ua);
  }else if(ep==="metagen"){
    reqUrl="/api/metagen?url="+encodeURIComponent(url);
    if(ua)reqUrl+="&ua="+encodeURIComponent(ua);
  }else if(ep==="version"){
    reqUrl="/version";
  }

  result.classList.add("visible");
  status.className="status";
  status.textContent="[TX] transmitting to "+reqUrl+" ...";
  body.textContent="";

  fetch(reqUrl).then(function(r){
    status.textContent="[RX] "+r.status+" "+r.statusText;
    if(!r.ok)status.classList.add("err");
    if(ep==="version")return r.text();
    if(ep==="metagen")return r.text();
    return r.json();
  }).then(function(data){
    if(ep==="version"){
      body.textContent=data;
    }else if(ep==="metagen"){
      body.textContent=data.substring(0,2000)+"\n... (truncated)";
    }else{
      var json=JSON.stringify(data,null,2);
      var imgHtml="";
      if(data.images&&data.images.length>0){
        imgHtml='<div class="img-panel"><details><summary>images ('+data.images.length+')</summary>';
        data.images.forEach(function(img){
          imgHtml+='<img src="'+img+'" alt="'+(data.title||'')+'">';
        });
        imgHtml+='</details></div>';
      }
      body.innerHTML='<pre>'+json.replace(/</g,"&lt;")+'</pre>'+imgHtml;
    }
  }).catch(function(e){
    status.className="status err";
    status.textContent="[ERR] "+e.message;
  });
}

document.getElementById("targetUrl").addEventListener("keydown",function(e){
  if(e.key==="Enter")sendRequest();
});
</script>

</body>
</html>`

// Utility functions

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

// -------- HEADLESS BROWSER FALLBACK --------

// Uses chromedp (Google Chrome) to fetch and parse metadata
func fetchMetaWithHeadless(target string, userAgent string) (string, string, []string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var htmlContent string

	err := chromedp.Run(ctx,
		emulation.SetUserAgentOverride(userAgent),
		chromedp.Navigate(target),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return "", "", nil, err
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", nil, err
	}

	title, desc, images := parseMetaFromHTML(doc, target)
	return title, desc, images, nil
}

// -------- METADATA EXTRACTION COMMON LOGIC --------

// Shared logic for extracting metadata from parsed HTML node
func parseMetaFromHTML(doc *html.Node, target string) (string, string, []string) {
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
				if text := extractText(n); text != "" && len(paragraphs) < 3 {
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

	if title == "" {
		if u, err := url.Parse(target); err == nil {
			title = u.Hostname()
		} else {
			title = "Untitled Page"
		}
	}

	if description == "" && len(paragraphs) > 0 {
		description = strings.Join(paragraphs, " ")
		if len(description) > 200 {
			description = description[:200] + "..."
		}
	} else if description == "" {
		description = "No description available"
	}

	if len(images) > 0 {
		baseURL, err := url.Parse(target)
		if err == nil {
			for i, img := range images {
				images[i] = fixURL(strings.TrimSpace(img), baseURL)
			}
		}
		images = uniqueStrings(images)
	}

	return title, description, images
}

// Main metadata fetch using HTTP client, fallback to headless browser if failed
func fetchMetaEnhanced(target string, ua string) (string, string, []string) {
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	switch ua {
	case "mobile":
		userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1"
	case "bot":
		userAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	case "twitter":
		userAgent = "Twitterbot/1.0"
	case "facebook":
		userAgent = "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"
	case "discord":
		userAgent = "Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return target, "", nil
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode == 403 || resp.StatusCode == 503) {
		// Fallback: Try fetching with headless browser
		title, desc, images, err2 := fetchMetaWithHeadless(target, userAgent)
		if err2 == nil {
			return title, desc, images
		}
		return target, "", nil
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		// Fallback: Try with headless browser if parsing fails
		title, desc, images, err2 := fetchMetaWithHeadless(target, userAgent)
		if err2 == nil {
			return title, desc, images
		}
		return target, "", nil
	}
	return parseMetaFromHTML(doc, target)
}

// Meta Extraction Utilities

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

// URL fixing helpers

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
