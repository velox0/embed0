    +----------------------------------------------------------------------------+
    |                                                                            |
    |        ___           ___           ___           ___           ___         |
    |       /\  \         /\__\         /\  \         /\  \         /\  \        |
    |      /::\  \       /::|  |       /::\  \       /::\  \       /::\  \       |
    |     /:/\:\  \     /:|:|  |      /:/\:\  \     /:/\:\  \     /:/\:\  \      |
    |    /::\~\:\  \   /:/|:|__|__   /::\~\:\__\   /::\~\:\  \   /:/  \:\__\     |
    |   /:/\:\ \:\__\ /:/ |::::\__\ /:/\:\ \:|__| /:/\:\ \:\__\ /:/__/ \:|__|    |
    |   \:\~\:\ \/__/ \/__/~~/:/  / \:\~\:\/:/  / \:\~\:\ \/__/ \:\  \ /:/  /    |
    |    \:\ \:\__\         /:/  /   \:\ \::/  /   \:\ \:\__\    \:\  /:/  /     |
    |     \:\ \/__/        /:/  /     \:\/:/  /     \:\ \/__/     \:\/:/  /      |
    |      \:\__\         /:/  /       \::/__/       \:\__\        \::/__/       |
    |       \/__/         \/__/         ~~            \/__/         ~~           |
    |                                                                            |
    |  metadata extraction & embed server                                        |
    |  v1.1.0                                                                    |
    |                                                                            |
    +----------------------------------------------------------------------------+

    ================================================================
      TABLE OF CONTENTS
    ================================================================

      1. Overview
      2. Quick Start
      3. Endpoints
      4. Embed Widget
      5. Configuration
      6. Architecture
      7. Documentation

    ================================================================
      1. OVERVIEW
    ================================================================

    embed0 is a Go-based metadata extraction server. Given any URL,
    it fetches the target page's Open Graph / HTML metadata and
    exposes that data through multiple interfaces:

      > JSON API         /api/embed
      > Meta Generator   /api/metagen
      > Embed Widget     /embed.js + /embed.css
      > API Explorer     / (homepage)

    Uses headless Chrome (chromedp) as fallback for JS-rendered
    or anti-bot protected pages.

    ================================================================
      2. QUICK START
    ================================================================

    Prerequisites:
      - Go 1.24+
      - Chrome/Chromium (for headless fallback)

    Setup:

      $ git clone https://github.com/velox0/embed0.git
      $ cd embed0
      $ cp .env.example .env
      $ go mod download
      $ go run app.go

    Server starts at http://localhost:8080

    Test:

      $ curl "http://localhost:8080/api/embed?url=https://github.com"
      $ curl "http://localhost:8080/api/metagen?url=https://example.com"

    ================================================================
      3. ENDPOINTS
    ================================================================

    GET /api/embed?url=<url>&theme=<theme>&ua=<ua>
    ---------------------------------------------------------
      Returns JSON metadata for a target URL.

      Response:
        {
          "title":       "GitHub",
          "description": "Let's build from here",
          "images":      ["https://..."],
          "url":         "https://github.com",
          "theme":       "light"
        }

    GET /api/metagen?url=<url>&ua=<ua>
    ---------------------------------------------------------
      Returns full HTML page with Open Graph and Twitter Card
      meta tags for the target URL.

    GET /embed.js
    ---------------------------------------------------------
      Minified JavaScript widget. Auto-detects server host.

    GET /embed.css
    ---------------------------------------------------------
      Stylesheet for embed widget. Three themes: light, dark,
      minimal.

    GET /version
    ---------------------------------------------------------
      Returns version string (plain text).

    GET /
    ---------------------------------------------------------
      Interactive API explorer homepage.

    ================================================================
      4. EMBED WIDGET
    ================================================================

    Drop-in link preview cards for any HTML page:

      <link rel="stylesheet" href="http://HOST/embed.css">
      <div data-url="https://example.com"></div>
      <div data-url="https://github.com" data-theme="dark"></div>
      <script async src="http://HOST/embed.js"></script>

    Attributes:
      data-url    (required) Target URL
      data-theme  (optional) light | dark | minimal

    The script fetches metadata from /api/embed and renders
    rich preview cards with images, title, description, and
    a link to the target.

    See docs/embed-widget.md for full details.

    ================================================================
      5. CONFIGURATION
    ================================================================

    Environment Variables (.env):

      HOST     Public URL of server       Default: http://localhost:8080
      PORT     Listen port                Default: 8080
      THEME    Default embed theme        Default: light

    ================================================================
      6. ARCHITECTURE
    ================================================================

    Request Flow:

      [Client]
          |
          v
      +--withCORS()----------+
      |                      |
      v                      |
    [Handler]                |
        |                    |
        v                    |
    [Fetcher]                |
        |                    |
        +-- HTTP GET ------->+-- Success? --> [Parser] --> Response
        |                    |
        +-- Fail/403/503 --->+-- Headless Chrome --> [Parser]

    Metadata Extraction Priority:

      Title:    og:title > <title> > hostname > "Untitled Page"
      Desc:     og:description > description > <h1>-<h3> > <p>
      Images:   og:image > twitter:image > <img> > /favicon.ico

    See docs/architecture.md for full system design.

    ================================================================
      7. DOCUMENTATION
    ================================================================

    Full docs in /docs:

      docs/api.md            API reference & examples
      docs/configuration.md  Environment variables & build
      docs/embed-widget.md   Widget integration guide
      docs/architecture.md   System design & data flow

    ================================================================
      LICENSE
    ================================================================

    MIT

    ================================================================
