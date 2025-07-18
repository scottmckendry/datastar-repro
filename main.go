package main

import (
	"bytes"
	"html/template"
	"net/http"
	"os"

	"github.com/starfederation/datastar-go/datastar"
)

var customers = []struct {
	ID   string
	Name string
}{
	{"1", "Alice"},
	{"2", "Bob"},
	{"3", "Charlie"},
}

var currentID = "1"

func main() {
	http.HandleFunc("/", handleIndex)

	// default morph - doesn't work
	http.HandleFunc("/sse/nav", handleNavSSE)

	// replace approach - works
	http.HandleFunc("/sse/nav-replace", handleNavReplaceSSE)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	println("Listening on :" + port)
	http.ListenAndServe(":"+port, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	id := currentID

	tmpl.Execute(w, struct {
		Customers []struct{ ID, Name string }
		CurrentID string
	}{customers, id})
}

func handleNavSSE(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}
	currentID = id

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("nav").Parse(navHTML))

	tmpl.Execute(buf, struct {
		Customers []struct{ ID, Name string }
		CurrentID string
	}{customers, id})

	sse := datastar.NewSSE(w, r)
	sse.PatchSignals([]byte(`{"currentID":"` + id + `"}`))
	sse.PatchElements(buf.String())
}

func handleNavReplaceSSE(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}

	currentID = id
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("nav").Parse(navHTML))

	tmpl.Execute(buf, struct {
		Customers []struct{ ID, Name string }
		CurrentID string
	}{customers, id})

	sse := datastar.NewSSE(w, r)
	sse.PatchSignals([]byte(`{"currentID":"` + id + `"}`))
	sse.PatchElements(buf.String(), datastar.WithModeReplace())
}

// change the datastar version between RC.1 and RC.2 to see the difference in behaviour
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Datastar Morph Repro</title>
	<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.1/bundles/datastar.js"></script>
	<style>
		.selected { background: #cce; }
	</style>
</head>
<body data-signals='{"currentID":"{{.CurrentID}}"}' data-on-datastar-sse='alert("Event Recieved")'>
	<ul id="nav-list">
		{{range .Customers}}
			<li>
				<a id="{{.ID}}" data-attr-class="$currentID == '{{.ID}}' ? 'selected' : ''">{{.Name}}</a>
			</li>
		{{end}}
	</ul>
	<div>
		{{range .Customers}}
			<button data-on-click="@get('/sse/nav?id={{.ID}}')">Default Morph for {{.Name}}</button>
			<button data-on-click="@get('/sse/nav-replace?id={{.ID}}')">Replace Morph for {{.Name}}</button>
		{{end}}
	</div>
	<div>
		<p>Click a name to select. In RC.1, this will generate an alert. In RC.2, it will not.</p>
	</div>
</body>
</html>
`

const navHTML = `{{define "nav"}}
<ul id="nav-list">
{{range .Customers}}
	<li>
		<a id="{{.ID}}" data-attr-class="$currentID == '{{.ID}}' ? 'selected' : ''">{{.Name}}</a>
	</li>
{{end}}
</ul>
{{end}}
`
