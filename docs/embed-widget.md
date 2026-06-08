# Embed Widget

    +-----------------------------------------------------+
    |  embed0 :: Embed Widget                              |
    |  Drop-in JS/CSS for link preview cards               |
    +-----------------------------------------------------+

================================================================
  QUICK START
================================================================

  1. Add the stylesheet:
  ---------------------------------------------------------------
  <link rel="stylesheet" href="http://localhost:8080/embed.css">

  2. Add target elements:
  ---------------------------------------------------------------
  <div data-url="https://example.com"></div>
  <div data-url="https://github.com" data-theme="dark"></div>

  3. Add the script:
  ---------------------------------------------------------------
  <script async src="http://localhost:8080/embed.js"></script>


================================================================
  HOW IT WORKS
================================================================

  Packet Flow:
  ---------------------------------------------------------------

  [Browser]                     [embed0 Server]
      |                               |
      |-- GET /api/embed?url=... ---->|
      |                               |-- HTTP GET target
      |                               |   (or headless Chrome)
      |                               |-- Parse meta tags
      |<-- { title, desc, images } ---|
      |                               |
      |-- Render embed card           |
      |                               |

  1. Script finds all [data-url] elements
  2. Fetches /api/embed for each URL
  3. Server fetches target page (HTTP or headless Chrome)
  4. Server returns parsed metadata as JSON
  5. Script renders rich preview card in the DOM


================================================================
  DATA ATTRIBUTES
================================================================

  Attribute   | Values              | Default
  ------------|---------------------|--------
  data-url    | Any valid URL       | (required)
  data-theme  | light | dark | minimal | light


================================================================
  RENDERED OUTPUT
================================================================

  Single Image:
  +----------------------------------------------------------+
  |  +------------------+  +------------------------------+  |
  |  |                  |  | Title of the Page            |  |
  |  |   [Image]        |  |                              |  |
  |  |                  |  | Description text goes here    |  |
  |  |                  |  | and can span multiple lines.  |  |
  |  +------------------+  |                              |  |
  |                        | Visit website                 |  |
  |                        +------------------------------+  |
  +----------------------------------------------------------+

  Multiple Images (2x2 grid):
  +----------------------------------------------------------+
  |  +------------------+  +------------------------------+  |
  |  | [img1] | [img2]  |  | Title of the Page            |  |
  |  |--------|---------|  |                              |  |
  |  | [img3] | [img4]  |  | Description text...          |  |
  |  +------------------+  | Visit website                 |  |
  |                        +------------------------------+  |
  +----------------------------------------------------------+


================================================================
  RESPONSIVE BREAKPOINTS
================================================================

  < 600px:   Stacked layout (image on top, content below)
  > 600px:   Side-by-side (40% image, 60% content)
  > 900px:   Max-width 800px container


================================================================
  ERROR HANDLING
================================================================

  If the fetch fails, the element displays:
    "Embed failed >_"

  Common failure modes:
    - Target URL unreachable
    - Server not running
    - CORS blocked (shouldn't happen with * origin)

