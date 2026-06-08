# Configuration

    +-----------------------------------------------------+
    |  embed0 :: Configuration                             |
    |  Environment Variables & Setup                       |
    +-----------------------------------------------------+

================================================================
  ENVIRONMENT VARIABLES
================================================================

  Variable  | Required | Default                  | Description
  ----------|----------|--------------------------|--------------------
  HOST      | No       | http://localhost:8080    | Public URL of the
            |          |                          | server (used by
            |          |                          | embed.js template)
  PORT      | No       | 8080                     | Listen port
  THEME     | No       | light                    | Default embed theme

----------------------------------------------------------------

  .env file location:  Project root (same directory as app.go)

  Example .env:
  --------------------------------------------------
  HOST=https://embed.example.com
  PORT=8080
  THEME=dark
  --------------------------------------------------


================================================================
  THEMES
================================================================

  Three themes available for the embed widget:

  light    - White background, dark text, blue links
  dark     - #1e1e1e background, white text, blue links
  minimal  - Transparent background, border-only, no shadow

  Set via ?theme= query param or THEME env var.


================================================================
  BUILD
================================================================

  Prerequisites:
    - Go 1.24+
    - Chrome/Chromium (for headless fallback)

  Build:
    $ go build -o app app.go

  Run:
    $ ./app

  Or without building:
    $ go run app.go


================================================================
  DOCKER
================================================================

  Not yet provided. See go.mod for dependencies.

  Key dependencies:
    - github.com/chromedp/chromedp    (headless Chrome)
    - github.com/joho/godotenv        (env loading)
    - github.com/tdewolff/minify/v2   (JS minification)
    - golang.org/x/net                (HTML parsing)

