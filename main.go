package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Options struct {
	ProjectName string
	Module      string
	Router      string
	Template    string
	CSS         string
	Auth        string
	PWA         bool
	Embed       bool
}

type fileTemplate struct {
	Path string
	Body string
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "create":
			createCommand(os.Args[2:])
		case "help", "--help", "-h":
			printUsage()
		default:
			fatal(fmt.Sprintf("unknown command %q\n\n%s", os.Args[1], usageText()))
		}
		return
	}
	createProject(bufio.NewReader(os.Stdin), "")
}

func createCommand(args []string) {
	if len(args) == 0 {
		fatal("usage: gothify create <project-name> [flags]")
	}
	projectName, flagArgs := args[0], args[1:]
	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	router := flags.String("router", "", "router: chi, fiber, echo, or net-http")
	frontend := flags.String("template", "", "frontend: templ, alpine, or inertia")
	css := flags.String("css", "", "css: tailwind, bulma, or plain")
	auth := flags.String("auth", "", "auth: session, jwt, or oauth2")
	module := flags.String("module", "", "Go module path")
	pwa := flags.Bool("pwa", false, "enable PWA support")
	embed := flags.Bool("embed", false, "embed static assets")
	if err := flags.Parse(flagArgs); err != nil {
		fatal(err.Error())
	}
	if *router == "" && *frontend == "" && *css == "" && *auth == "" && *module == "" && !*pwa && !*embed {
		createProject(bufio.NewReader(os.Stdin), projectName)
		return
	}
	options, err := optionsFromFlags(projectName, *router, *frontend, *css, *auth, *module, *pwa, *embed)
	if err != nil {
		fatal(err.Error())
	}
	if err := generate(options); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("\nDone. Project generated at ./%s\n", projectName)
}

func optionsFromFlags(projectName, router, frontend, css, auth, module string, pwa, embed bool) (Options, error) {
	if !validProjectName(projectName) {
		return Options{}, fmt.Errorf("project directory must use letters, numbers, dashes, or underscores")
	}
	lookup := func(value string, values map[string]string, label string) (string, error) {
		result, ok := values[strings.ToLower(value)]
		if !ok {
			return "", fmt.Errorf("invalid %s %q", label, value)
		}
		return result, nil
	}
	selectedRouter, err := lookup(router, map[string]string{"chi": "Chi (lightweight and idiomatic Go)", "fiber": "Fiber (Express-like and fast)", "echo": "Echo (feature-complete and structured)", "net-http": "Go standard library (net/http)"}, "router")
	if err != nil {
		return Options{}, err
	}
	selectedFrontend, err := lookup(frontend, map[string]string{"templ": "Go Templ + HTMX (native Go server-side rendering)", "alpine": "Go html/template + Alpine.js (standard Go templating)", "inertia": "Go + Inertia.js (Go monolith with React/Vue)"}, "template")
	if err != nil {
		return Options{}, err
	}
	selectedCSS, err := lookup(css, map[string]string{"tailwind": "Tailwind CSS", "bulma": "Bulma / PicoCSS", "plain": "None / Plain CSS"}, "css")
	if err != nil {
		return Options{}, err
	}
	selectedAuth, err := lookup(auth, map[string]string{"session": "Session-Based (Cookie)", "jwt": "JWT + Refresh Token", "oauth2": "OAuth2 (Google / GitHub via Goth)"}, "auth")
	if err != nil {
		return Options{}, err
	}
	if module == "" {
		module = "github.com/fahmi-azzuhri/" + projectName
	}
	return Options{ProjectName: projectName, Module: module, Router: selectedRouter, Template: selectedFrontend, CSS: selectedCSS, Auth: selectedAuth, PWA: pwa, Embed: embed}, nil
}

func createProject(reader *bufio.Reader, requestedName string) {
	fmt.Println("Gothify - Fullstack Go Modern generator")
	fmt.Println("Generate a pragmatic Go MPA/SPA without a heavy JavaScript framework.")
	projectName := requestedName
	if projectName == "" {
		projectName = ask(reader, "Project directory", "my-goth-app")
	}
	if !validProjectName(projectName) {
		fatal("project directory must use letters, numbers, dashes, or underscores")
	}
	routerChoice := choose(reader, "Go HTTP router / framework", []string{"Chi (lightweight and idiomatic Go)", "Fiber (Express-like and fast)", "Echo (feature-complete and structured)", "Go standard library (net/http)"})
	templateChoice := choose(reader, "Template engine & interactivity", []string{"Go Templ + HTMX (native Go server-side rendering)", "Go html/template + Alpine.js (standard Go templating)", "Go + Inertia.js (Go monolith with React/Vue)"})
	cssChoice := choose(reader, "CSS framework", []string{"Tailwind CSS", "Bulma / PicoCSS", "None / Plain CSS"})
	authChoice := choose(reader, "Authentication strategy", []string{"Session-Based (Cookie)", "JWT + Refresh Token", "OAuth2 (Google / GitHub via Goth)"})
	pwa := confirm(reader, "Add PWA support")
	embed := confirm(reader, "Embed static assets with go:embed")
	module := "github.com/fahmi-azzuhri/" + projectName
	options := Options{ProjectName: projectName, Module: module, Router: routerChoice, Template: templateChoice, CSS: cssChoice, Auth: authChoice, PWA: pwa, Embed: embed}
	if err := generate(options); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("\nDone. Project generated at ./%s\n", projectName)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n  go run ./cmd/server\n", projectName)
}

func usageText() string {
	return `Usage:
  gothify create <project-name>   Create an interactive project
  gothify create <name> [flags]   Create without prompts
  gothify help                    Show this help

Examples:
  gothify create my-app
  gothify create my-app --router chi --template templ --css tailwind --auth session --pwa --embed
  gothify create store-backend --router echo --template alpine --css plain --auth jwt`
}

func printUsage() { fmt.Println("Gothify - Fullstack Go Modern generator"); fmt.Println(usageText()) }

func ask(reader *bufio.Reader, label, fallback string) string {
	fmt.Printf("%s [%s]: ", label, fallback)
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func choose(reader *bufio.Reader, label string, choices []string) string {
	fmt.Printf("\n%s\n", label)
	for index, choice := range choices {
		fmt.Printf("  %d. %s\n", index+1, choice)
	}
	for {
		value := ask(reader, "Choose", "1")
		index, err := strconv.Atoi(value)
		if err == nil && index >= 1 && index <= len(choices) {
			return choices[index-1]
		}
		fmt.Printf("Please choose a number from 1 to %d.\n", len(choices))
	}
}

func confirm(reader *bufio.Reader, label string) bool {
	for {
		value := strings.ToLower(ask(reader, label+"? (y/n)", "n"))
		switch value {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please answer y or n.")
		}
	}
}
