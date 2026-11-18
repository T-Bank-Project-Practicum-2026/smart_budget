// Package swagger serves the generated OpenAPI document and Swagger UI.
package swagger

import (
	_ "embed"
	"net/http"

	swaggerFiles "github.com/swaggo/files/v2"
)

//go:embed openapi.yaml
var openAPIDocument []byte

const indexHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Smart Budget API</title>
  <link rel="stylesheet" href="/swagger/swagger-ui.css">
  <link rel="icon" type="image/png" href="/swagger/favicon-32x32.png" sizes="32x32">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="/swagger/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="/swagger/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/swagger/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        docExpansion: "list",
        withCredentials: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        plugins: [SwaggerUIBundle.plugins.DownloadUrl],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>
`

// Handler serves Swagger UI under /swagger/ and the OpenAPI document under
// /swagger/openapi.yaml.
func Handler() http.Handler {
	assets := http.StripPrefix("/swagger/", http.FileServer(http.FS(swaggerFiles.FS)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger/", "/swagger/index.html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(indexHTML))

		case "/swagger/openapi.yaml":
			w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
			_, _ = w.Write(openAPIDocument)

		default:
			assets.ServeHTTP(w, r)
		}
	})
}
