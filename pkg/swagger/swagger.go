package swagger

import (
	_ "embed"

	"github.com/gofiber/fiber/v3"
)

//go:embed openapi.yaml
var openAPISpec string

func Register(app *fiber.App) {
	app.Get("/swagger", func(c fiber.Ctx) error {
		c.Type("html")
		return c.SendString(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Job Prep Backend Swagger</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
	<style>
		body { margin: 0; background: #0f172a; }
		#swagger-ui { min-height: 100vh; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = () => {
			window.ui = SwaggerUIBundle({
				url: '/swagger/openapi.yaml',
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
				layout: 'BaseLayout'
			});
		};
	</script>
</body>
</html>`)
	})

	app.Get("/swagger/openapi.yaml", func(c fiber.Ctx) error {
		c.Type("yaml")
		return c.SendString(openAPISpec)
	})
}
