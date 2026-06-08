# API Reference

    +-----------------------------------------------------+
    |  embed0 :: API Reference                             |
    |  Protocol: HTTP/1.1 | Format: JSON                   |
    +-----------------------------------------------------+

================================================================
  ENDPOINTS
================================================================

  GET /api/embed
  ---------------------------------------------------------------
  Query Parameters:

    url     (required)  Target URL to extract metadata from
    theme   (optional)  Widget theme: light | dark | minimal
    ua      (optional)  User-Agent preset:
                          - mobile
                          - bot
                          - twitter
                          - facebook
                          - discord

  Response (200 OK):
  {
    "title":       "string",
    "description": "string",
    "images":      ["string"],
    "url":         "string",
    "theme":       "string"
  }

  Errors:
    400  Missing ?url parameter
    500  Failed to fetch target URL

----------------------------------------------------------------

  GET /api/metagen
  ---------------------------------------------------------------
  Query Parameters:

    url     (required)  Target URL to extract metadata from
    ua      (optional)  User-Agent preset (same as /api/embed)

  Response (200 OK):
    Full HTML document with Open Graph and Twitter Card meta tags.

  Errors:
    400  Missing URL in path
    400  Invalid URL encoding

----------------------------------------------------------------

  GET /embed.js
  ---------------------------------------------------------------
  Response (200 OK):
    JavaScript widget (minified). Content-Type: application/javascript

----------------------------------------------------------------

  GET /embed.css
  ---------------------------------------------------------------
  Response (200 OK):
    Stylesheet for embed widget. Content-Type: text/css

----------------------------------------------------------------

  GET /version
  ---------------------------------------------------------------
  Response (200 OK):
    Plain text version string. Content-Type: text/plain

----------------------------------------------------------------

  GET /
  ---------------------------------------------------------------
  Response (200 OK):
    Interactive API explorer / homepage. Content-Type: text/html


================================================================
  CORS
================================================================

  All endpoints include CORS headers:

    Access-Control-Allow-Origin:  *
    Access-Control-Allow-Methods: GET, OPTIONS
    Access-Control-Allow-Headers: Content-Type

  OPTIONS requests return 200 OK immediately.


================================================================
  USER-AGENT PRESETS
================================================================

  Preset       | User-Agent String
  -------------|--------------------------------------------------
  (default)    | Mozilla/5.0 (Windows NT 10.0; Win64; x64) ...
  mobile       | Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 ...)
  bot          | Mozilla/5.0 (compatible; Googlebot/2.1; ...)
  twitter      | Twitterbot/1.0
  facebook     | facebookexternalhit/1.1
  discord      | Mozilla/5.0 (compatible; Discordbot/2.0; ...)


================================================================
  RATE LIMITS
================================================================

  None enforced. Client-side HTTP timeout: 10 seconds.
  Headless browser fallback timeout: ~3 seconds page load.


================================================================
  EXAMPLES
================================================================

  # Fetch metadata for github.com
  curl "http://localhost:8080/api/embed?url=https://github.com"

  # Fetch with dark theme preset
  curl "http://localhost:8080/api/embed?url=https://example.com&theme=dark"

  # Fetch as Twitterbot
  curl "http://localhost:8080/api/embed?url=https://example.com&ua=twitter"

  # Generate meta tags HTML
  curl "http://localhost:8080/api/metagen?url=https://example.com"

  # Get version
  curl "http://localhost:8080/version"

