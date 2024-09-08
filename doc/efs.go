package docs

import "embed"

//go:embed "api/api_v1.yaml"
var ApiV1 embed.FS

//go:embed "api/swagger.html"
var SwaggerUI embed.FS
