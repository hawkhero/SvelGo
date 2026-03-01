package svelgo

import "html/template"

type pageRenderData struct {
	PageID      string
	StateBlob   string
	Manifest    template.JS // raw JSON — not HTML-escaped
	AssetScript string
	AssetCSS    string
	Debug       bool
}

var shellTemplate = template.Must(template.New("shell").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>SvelGo</title>
  {{- if .AssetCSS}}
  <link rel="stylesheet" href="{{.AssetCSS}}">
  {{- end}}
</head>
<body>
  <script>
    window.__SVELGO_PAGE_ID__  = "{{.PageID}}";
    window.__SVELGO_STATE__    = "{{.StateBlob}}";
    window.__SVELGO_MANIFEST__ = {{.Manifest}};
    window.__SVELGO_DEBUG__    = {{.Debug}};
  </script>
  <div id="svelgo-root"></div>
  <script type="module" src="{{.AssetScript}}"></script>
</body>
</html>
`))
