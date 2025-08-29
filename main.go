package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// Config item definition
type ConfigItem struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon,omitempty"`
	Min   int    `json:"min,omitempty"`
	Max   int    `json:"max,omitempty"`
	URL   string `json:"url"`
	Value int    // default value for sliders
}

var config []ConfigItem
var tmpl *template.Template

func main() {
	// Load config.json
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config.json: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error parsing config.json: %v", err)
	}

	// Fetch default slider values
	for i := range config {
		if config[i].Type == "slider" {
			config[i].Value = fetchDefaultValue(config[i].URL)
		}
	}

	// Load template
	tmpl = template.Must(template.ParseFiles("templates/index.html"))

	// Routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/activate", handleActivate)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Simulated call to backend API to fetch slider defaults
func fetchDefaultValue(url string) int {
	// Example: call API (fake implementation for now)
	// resp, err := http.Get(url + "/default")
	// ... parse response
	return 50 // placeholder default value
}

// Render HTML with config
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handle frontend actions and call mapped URLs
func handleActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name  string `json:"name"`
		Value string `json:"value,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Find item in config
	var item *ConfigItem
	for i, c := range config {
		if c.Name == req.Name {
			item = &config[i]
			break
		}
	}
	if item == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Construct target URL
	finalURL := item.URL
	if item.Type == "slider" {
		finalURL = fmt.Sprintf("%s?value=%s", item.URL, req.Value)
	}

	// Call the mapped URL
	resp, err := http.Get(finalURL)
	if err != nil {
		http.Error(w, "Failed to call target URL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Triggered %s", item.Name)
}
