package main

import "fmt"

func serverMainTemplate(options Options) string {
	switch options.Router {
	case "Chi (lightweight and idiomatic Go)":
		return chiServerTemplate(options)
	case "Fiber (Express-like and fast)":
		return fiberServerTemplate(options)
	case "Echo (feature-complete and structured)":
		return echoServerTemplate(options)
	}
	assetImport := ""
	fsImport := ""
	assetSetup := `
	staticFS := http.Dir("web/static")
`
	assetHandler := `http.FileServer(staticFS)`
	if options.Embed {
		assetImport = fmt.Sprintf("\n\t\"%s/web\"", options.Module)
		fsImport = "\n\t\"io/fs\""
		assetSetup = `
	embeddedFS, err := fs.Sub(web.Files, "static")
	if err != nil {
		log.Fatal(err)
	}
	staticFS := http.FS(embeddedFS)
`
	}
	return fmt.Sprintf(`package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"%s%s
)

func main() {
	page, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil { log.Fatal(err) }
%s
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", %s))
	mux.HandleFunc("GET /health", func(writer http.ResponseWriter, request *http.Request) { writer.Write([]byte("ok")) })
	%s
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(writer, nil); err != nil { http.Error(writer, "template rendering failed", http.StatusInternalServerError) }
	})
	log.Println("server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
`, fsImport, assetImport, assetSetup, assetHandler, frontendRoute(options, "mux"))
}

func frontendRoute(options Options, routerName string) string {
	if options.Template == "Go + Inertia.js (Go monolith with React/Vue)" && routerName == "mux" {
		return `mux.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("web/frontend"))))`
	}
	return ""
}

func assetTemplateParts(options Options) (string, string, string) {
	assetImport := ""
	fsImport := ""
	assetSetup := `staticFS := http.Dir("web/static")`
	if options.Embed {
		assetImport = fmt.Sprintf("\n\t\"%s/web\"", options.Module)
		fsImport = "\n\t\"io/fs\""
		assetSetup = `embeddedFS, err := fs.Sub(web.Files, "static")
	if err != nil { log.Fatal(err) }
	staticFS := http.FS(embeddedFS)`
	}
	return assetImport, fsImport, assetSetup
}

func chiServerTemplate(options Options) string {
	assetImport, fsImport, assetSetup := assetTemplateParts(options)
	return fmt.Sprintf(`package main

import (
	"html/template"
	"log"
	"net/http"%s
	"path/filepath"%s
	"github.com/go-chi/chi/v5"
)

func main() {
	page, err := template.ParseFiles(filepath.Join("web", "templates", "index.html")); if err != nil { log.Fatal(err) }
	%s
	router := chi.NewRouter()
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticFS)))
	router.Get("/health", func(writer http.ResponseWriter, request *http.Request) { writer.Write([]byte("ok")) })
	router.Get("/", func(writer http.ResponseWriter, request *http.Request) { writer.Header().Set("Content-Type", "text/html; charset=utf-8"); _ = page.Execute(writer, nil) })
	log.Println("server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
`, fsImport, assetImport, assetSetup)
}

func fiberServerTemplate(options Options) string {
	assetImport, fsImport, assetSetup := assetTemplateParts(options)
	return fmt.Sprintf(`package main

import (
	"html/template"
	"log"
	"net/http"%s
	"path/filepath"%s
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func main() {
	page, err := template.ParseFiles(filepath.Join("web", "templates", "index.html")); if err != nil { log.Fatal(err) }
	%s
	app := fiber.New()
	app.Use("/static", filesystem.New(filesystem.Config{Root: staticFS}))
	app.Get("/health", func(context *fiber.Ctx) error { return context.SendString("ok") })
	app.Get("/", func(context *fiber.Ctx) error { context.Set(fiber.HeaderContentType, "text/html; charset=utf-8"); return page.Execute(context, nil) })
	log.Println("server listening on http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
`, fsImport, assetImport, assetSetup)
}

func echoServerTemplate(options Options) string {
	assetImport, fsImport, assetSetup := assetTemplateParts(options)
	return fmt.Sprintf(`package main

import (
	"html/template"
	"io"
	"log"
	"net/http"%s
	"path/filepath"%s
	"github.com/labstack/echo/v4"
)

type renderer struct { page *template.Template }
func (renderer renderer) Render(writer io.Writer, name string, data interface{}, context echo.Context) error { return renderer.page.Execute(writer, data) }

func main() {
	page, err := template.ParseFiles(filepath.Join("web", "templates", "index.html")); if err != nil { log.Fatal(err) }
	%s
	echoServer := echo.New(); echoServer.Renderer = renderer{page: page}
	echoServer.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(staticFS))))
	echoServer.GET("/health", func(context echo.Context) error { return context.String(http.StatusOK, "ok") })
	echoServer.GET("/", func(context echo.Context) error { return context.Render(http.StatusOK, "index.html", nil) })
	log.Println("server listening on http://localhost:8080")
	log.Fatal(echoServer.Start(":8080"))
}
`, fsImport, assetImport, assetSetup)
}

func routerTemplate(router string) string {
	return fmt.Sprintf(`package httpserver

// Router selection: %s. This package documents the selected HTTP adapter.
type Router string

const Selected Router = %q
`, router, router)
}
