package urls

import (
	"bytes"
	"html/template"

	"github.com/gofiber/fiber/v2"
)

const tmpl = `
{{ range $name, $routes := . }}
	<h1>{{ $name }} </h1>
	<ul>
	{{ range $routes }}
		<li>{{ .Method }} - {{ .Path }}</li>
	{{ end }}
	</ul>
{{ end }}
`

func getUrls(u *urls) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		routes := u.pluginApi.GetAppRoutes()
		routes["Project"] = u.pluginApi.GetProjectRoutes()

		buf := &bytes.Buffer{}
		t := template.Must(template.New("tmpl").Parse(tmpl))
		if err := t.Execute(buf, routes); err != nil {
			return err
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(buf.String())
	}
}
