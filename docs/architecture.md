# Architecture

    +-----------------------------------------------------+
    |  embed0 :: Architecture                              |
    |  System Design & Data Flow                           |
    +-----------------------------------------------------+

================================================================
  COMPONENTS
================================================================

  +----------------------------------------------------------+
  |                      embed0 Server                       |
  |                                                          |
  |  +----------+  +----------+  +----------+  +----------+ |
  |  | Homepage |  | API      |  | Widget   |  | MetaGen  | |
  |  | Handler  |  | Handler  |  | JS/CSS   |  | Handler  | |
  |  | /        |  | /api/*   |  | /embed.* |  | /api/metagen | |
  |  +----+-----+  +----+-----+  +----+-----+  +----+-----+ |
  |       |              |              |              |      |
  |       |         +----v----+         |              |      |
  |       |         | Fetcher |         |              |      |
  |       |         | Engine  |         |              |      |
  |       |         +----+----+         |              |      |
  |       |              |              |              |      |
  |       |         +----v----+         |              |      |
  |       |         | Parser  |         |              |      |
  |       |         +----+----+         |              |      |
  |       |              |              |              |      |
  |       |         +----v------+       |              |      |
  |       |         | Headless  |       |              |      |
  |       |         | Chrome    |       |              |      |
  |       |         | Fallback  |       |              |      |
  |       |         +-----------+       |              |      |
  +----------------------------------------------------------+


================================================================
  REQUEST FLOW
================================================================

  1. Client Request
     --------------------------------------------------------
     GET /api/embed?url=https://example.com HTTP/1.1
     Host: localhost:8080

  2. CORS Middleware (withCORS)
     --------------------------------------------------------
     Sets headers, handles OPTIONS preflight

  3. Handler (embedHandler)
     --------------------------------------------------------
     Validates ?url param, resolves THEME default

  4. Fetcher (fetchMetaEnhanced)
     --------------------------------------------------------
     +---> Try HTTP GET with User-Agent
     |     |
     |     +---> Success? ---> Parse HTML ---> Return
     |     |
     |     +---> 403/503 or error?
     |           |
     |           +---> Headless Chrome (chromedp)
     |                 |
     |                 +---> Navigate, wait 3s, extract HTML
     |                 |
     |                 +---> Parse HTML ---> Return
     |

  5. Parser (parseMetaFromHTML)
     --------------------------------------------------------
     Walks DOM tree extracting:
       - <title>
       - <meta> tags (og:title, og:description, og:image,
                      twitter:image, description)
       - <h1>, <h2>, <h3> as fallback description
       - <p> tags (first 3) as fallback description
       - <img> tags as fallback image

  6. Response
     --------------------------------------------------------
     JSON encoded PageInfo struct returned to client


================================================================
  METADATA EXTRACTION PRIORITY
================================================================

  Title:
    1. og:title meta tag
    2. <title> element
    3. Hostname of target URL
    4. "Untitled Page"

  Description:
    1. og:description meta tag
    2. description meta tag
    3. First <h1>-<h3> text
    4. First 3 <p> texts joined (truncated to 200 chars)
    5. "No description available"

  Images:
    1. og:image meta tag(s)
    2. twitter:image meta tag(s)
    3. First <img> src on page
    4. /favicon.ico fallback

  All relative URLs are resolved against the target URL's origin.


================================================================
  HEADLESS FALLBACK
================================================================

  Triggered when:
    - HTTP request returns 403 or 503
    - HTTP request fails entirely
    - HTML parsing fails

  Behavior:
    - Launches chromedp browser context
    - Sets User-Agent to match request
    - Navigates to target URL
    - Waits 3 seconds for JS rendering
    - Extracts full outerHTML of <html>
    - Parses same way as HTTP response

  Timeout: Governed by chromedp defaults (~30s total)


================================================================
  FILE STRUCTURE
================================================================

  embed0/
  |-- app.go              Main server (all handlers + logic)
  |-- go.mod              Module definition
  |-- go.sum              Dependency checksums
  |-- .env                Runtime config
  |-- .env.example        Config template
  |-- example.html        Demo page
  |-- static/
  |   |-- embed.js        Widget script (source)
  |   |-- embed.css       Widget stylesheet
  |-- docs/
      |-- api.md          API reference
      |-- configuration.md
      |-- embed-widget.md
      |-- architecture.md

