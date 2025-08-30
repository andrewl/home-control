package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// Control represents one UI element (button or slider)
type Control struct {
	Name string `json:"name"`
	Type        string `json:"type"` // "button" or "slider"
	Icon        string `json:"icon"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
	URL         string `json:"url"`
	Value       int    // dynamically set from status API
}

// Config represents the entire app config
type Config struct {
	StatusURL string    `json:"status_url"`
	Controls  []Control `json:"controls"`
}

// HTTP client with a sane timeout
var httpClient = &http.Client{Timeout: 5 * time.Second}

// loadConfig reads config.json fresh from disk
func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// fetchCurrentValues calls the status API and returns a map[id]value
func fetchCurrentValues(url string) (map[string]int, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var values []struct {
		ID    string `json:"id"`
		Value int    `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil, err
	}

	result := make(map[string]int, len(values))
	for _, v := range values {
		result[v.ID] = v.Value
	}
	return result, nil
}

// renderHandler renders the UI
func renderHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		http.Error(w, "Config load error" + err.Error()	, http.StatusInternalServerError)
		return
	}

	// Fetch slider values (best-effort)
	if cfg.StatusURL != "" {
		if values, err := fetchCurrentValues(cfg.StatusURL); err != nil {
			log.Printf("fetchCurrentValues error: %v", err)
		} else {
			for i := range cfg.Controls {
				if val, ok := values[cfg.Controls[i].Name]; ok {
					cfg.Controls[i].Value = val
				}
			}
		}
	}

	// Parse template each time to allow live edits (dev-friendly)
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, cfg.Controls); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// activateHandler is called when a button is pressed or slider moved
func activateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Printf("Config load error: %v", err)
		http.Error(w, "Config load error", http.StatusInternalServerError)
		return
	}

	var req struct {
		Name    string `json:"name"`
		Value string    `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Bad request body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	log.Printf("Activate request: %+v", req)

	// Find control
	var target *Control
	for i := range cfg.Controls {
		if cfg.Controls[i].Name == req.Name {
			log.Printf("Found control: %+v", cfg.Controls[i])
			target = &cfg.Controls[i]
			break
		}
	}
	if target == nil {
		log.Printf("Unknown control: %s", req.Name)
		http.Error(w, "unknown control", http.StatusBadRequest)
		return
	}

	// Build outbound request
	var (
		outReq *http.Request
	)
	if target.Type == "slider" {
		body, _ := json.Marshal(map[string]string{"value": req.Value})
		outReq, err = http.NewRequest(http.MethodPost, target.URL, bytes.NewReader(body))
		outReq.Header.Set("Content-Type", "application/json")
	} else {
		outReq, err = http.NewRequest(http.MethodGet, target.URL, bytes.NewReader([]byte("{}")))
		outReq.Header.Set("Content-Type", "application/json")
	}
	if err != nil {
		http.Error(w, "request build failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	log.Printf("Calling backend: %s %s", outReq.Method, target.URL)

	resp, err := httpClient.Do(outReq)
	if err != nil {
		http.Error(w, "backend call failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	log.Printf("Backend response: %s", resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		http.Error(w, "backend returned error: "+resp.Status + " - " + target.URL, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"status": "success",
	})
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", renderHandler)
	http.HandleFunc("/activate", activateHandler)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

