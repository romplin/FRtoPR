package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type FeatureRequest struct {
	ID                   int       `json:"id"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	AcceptanceCriteria   string    `json:"acceptance_criteria"`
	Priority             string    `json:"priority"`
	TargetTimeline       string    `json:"target_timeline"`
	AffectedComponents   []string  `json:"affected_components"`
	ExampleUsage         string    `json:"example_usage"`
	TechnicalConstraints string    `json:"technical_constraints"`
	CreatedAt            time.Time `json:"created_at"`
	Status               string    `json:"status"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var featureRequests []FeatureRequest
var nextID = 1

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/submit", handleSubmit)
	http.HandleFunc("/api/requests", handleRequests)
	http.HandleFunc("/health", handleHealth)
	
	// New HTMX-specific endpoints
	http.HandleFunc("/htmx/submit", handleHTMXSubmit)
	http.HandleFunc("/htmx/requests", handleHTMXRequests)
	http.HandleFunc("/htmx/form", handleHTMXForm)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Feature Request System</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #007bff;
            padding-bottom: 10px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
            color: #333;
        }
        input, textarea, select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
            box-sizing: border-box;
        }
        textarea {
            resize: vertical;
            min-height: 100px;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            margin-right: 10px;
        }
        button:hover {
            background-color: #0056b3;
        }
        button:disabled {
            background-color: #6c757d;
            cursor: not-allowed;
        }
        .secondary-btn {
            background-color: #6c757d;
        }
        .success-message {
            background-color: #d4edda;
            color: #155724;
            padding: 15px;
            border-radius: 4px;
            margin-bottom: 20px;
            border: 1px solid #c3e6cb;
        }
        .error-message {
            background-color: #f8d7da;
            color: #721c24;
            padding: 15px;
            border-radius: 4px;
            margin-bottom: 20px;
            border: 1px solid #f5c6cb;
        }
        .feature-item {
            background: #f8f9fa;
            padding: 20px;
            margin-bottom: 15px;
            border-radius: 4px;
            border-left: 4px solid #007bff;
        }
        .feature-title {
            font-size: 18px;
            font-weight: 600;
            color: #333;
            margin-bottom: 10px;
        }
        .feature-meta {
            font-size: 12px;
            color: #666;
            margin-bottom: 10px;
        }
        .tabs {
            display: flex;
            margin-bottom: 20px;
            border-bottom: 1px solid #ddd;
        }
        .tab {
            padding: 10px 20px;
            cursor: pointer;
            border-bottom: 2px solid transparent;
            color: #666;
            background: none;
            border-left: none;
            border-right: none;
            border-top: none;
            border-radius: 0;
            margin-right: 0;
        }
        .tab.active {
            color: #007bff;
            border-bottom-color: #007bff;
        }
        .htmx-request {
            opacity: 0.5;
        }
        .content-section {
            display: none;
        }
        .content-section.active {
            display: block;
        }
        .loading {
            text-align: center;
            padding: 20px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Feature Request System</h1>
        
        <div class="tabs">
            <button 
                class="tab active" 
                onclick="showTab('form')"
                id="form-tab"
            >
                New Request
            </button>
            <button 
                class="tab" 
                onclick="showTab('list')"
                id="list-tab"
                hx-get="/htmx/requests"
                hx-target="#requests-content"
                hx-trigger="click"
            >
                All Requests
            </button>
        </div>

        <div id="form-content" class="content-section active">
            <div id="message-container"></div>
            
            <form 
                hx-post="/htmx/submit" 
                hx-target="#message-container"
                hx-swap="innerHTML"
                hx-on::after-request="if(event.detail.successful) { document.getElementById('feature-form').reset(); }"
                id="feature-form"
            >
                <div class="form-group">
                    <label for="title">Feature Title *</label>
                    <input 
                        type="text" 
                        id="title" 
                        name="title"
                        placeholder="Brief descriptive title for the feature"
                        required
                    >
                </div>

                <div class="form-group">
                    <label for="description">Feature Description *</label>
                    <textarea 
                        id="description" 
                        name="description"
                        placeholder="Detailed description of the feature requirements and functionality"
                        required
                    ></textarea>
                </div>

                <div class="form-group">
                    <label for="acceptance_criteria">Acceptance Criteria *</label>
                    <textarea 
                        id="acceptance_criteria" 
                        name="acceptance_criteria"
                        placeholder="Clear, testable criteria that define when this feature is complete"
                        required
                    ></textarea>
                </div>

                <div class="form-group">
                    <label for="priority">Priority Level *</label>
                    <select 
                        id="priority" 
                        name="priority"
                        required
                    >
                        <option value="">Select priority</option>
                        <option value="high">High - Critical/Urgent</option>
                        <option value="medium">Medium - Important</option>
                        <option value="low">Low - Nice to have</option>
                    </select>
                </div>

                <div class="form-group">
                    <label for="target_timeline">Target Timeline</label>
                    <input 
                        type="text" 
                        id="target_timeline" 
                        name="target_timeline"
                        placeholder="e.g., Next Sprint, Q2 2025, 2 weeks"
                    >
                </div>

                <div class="form-group">
                    <label for="affected_components">Affected Components/Modules</label>
                    <input 
                        type="text" 
                        id="affected_components" 
                        name="affected_components"
                        placeholder="api, frontend, database, auth-service"
                    >
                </div>

                <div class="form-group">
                    <label for="example_usage">Example Usage Scenarios</label>
                    <textarea 
                        id="example_usage" 
                        name="example_usage"
                        placeholder="Provide specific examples of how users would interact with this feature"
                    ></textarea>
                </div>

                <div class="form-group">
                    <label for="technical_constraints">Technical Constraints or Preferences</label>
                    <textarea 
                        id="technical_constraints" 
                        name="technical_constraints"
                        placeholder="Any technical limitations, preferred technologies, or implementation constraints"
                    ></textarea>
                </div>

                <button type="submit">Submit Feature Request</button>
                <button type="button" onclick="document.getElementById('feature-form').reset();" class="secondary-btn">Reset Form</button>
            </form>
        </div>

        <div id="list-content" class="content-section">
            <div id="requests-content">
                <div class="loading">Click "All Requests" to load feature requests...</div>
            </div>
        </div>
    </div>

    <script>
        function showTab(tabName) {
            // Hide all content sections
            document.querySelectorAll('.content-section').forEach(section => {
                section.classList.remove('active');
            });
            
            // Remove active class from all tabs
            document.querySelectorAll('.tab').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // Show selected content and activate tab
            document.getElementById(tabName + '-content').classList.add('active');
            document.getElementById(tabName + '-tab').classList.add('active');
        }

        // Auto-hide success messages after 5 seconds
        document.body.addEventListener('htmx:afterSwap', function(event) {
            if (event.detail.target.id === 'message-container') {
                const messageEl = event.detail.target.querySelector('.success-message');
                if (messageEl) {
                    setTimeout(() => {
                        messageEl.remove();
                    }, 5000);
                }
            }
        });
    </script>
</body>
</html>`
	fmt.Fprint(w, html)
}

func handleHTMXSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `<div class="error-message">Method not allowed</div>`)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `<div class="error-message">Error parsing form data</div>`)
		return
	}

	// Parse affected components
	var components []string
	if componentsStr := r.FormValue("affected_components"); componentsStr != "" {
		for _, comp := range strings.Split(componentsStr, ",") {
			comp = strings.TrimSpace(comp)
			if comp != "" {
				components = append(components, comp)
			}
		}
	}

	featureRequest := FeatureRequest{
		ID:                   nextID,
		Title:                r.FormValue("title"),
		Description:          r.FormValue("description"),
		AcceptanceCriteria:   r.FormValue("acceptance_criteria"),
		Priority:             r.FormValue("priority"),
		TargetTimeline:       r.FormValue("target_timeline"),
		AffectedComponents:   components,
		ExampleUsage:         r.FormValue("example_usage"),
		TechnicalConstraints: r.FormValue("technical_constraints"),
		CreatedAt:            time.Now(),
		Status:               "submitted",
	}

	// Validate required fields
	if featureRequest.Title == "" || featureRequest.Description == "" ||
		featureRequest.AcceptanceCriteria == "" || featureRequest.Priority == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `<div class="error-message">Please fill in all required fields</div>`)
		return
	}

	// Save the feature request
	featureRequests = append(featureRequests, featureRequest)
	nextID++

	// Return success message
	fmt.Fprintf(w, `<div class="success-message">Feature request submitted successfully! ID: %d</div>`, featureRequest.ID)
}

func handleHTMXRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `<div class="error-message">Method not allowed</div>`)
		return
	}

	if len(featureRequests) == 0 {
		fmt.Fprint(w, `<div class="feature-item"><div>No feature requests found.</div></div>`)
		return
	}

	// Generate HTML for all feature requests
	for _, request := range featureRequests {
		fmt.Fprintf(w, `
		<div class="feature-item">
			<div class="feature-title">%s</div>
			<div class="feature-meta">
				ID: %d | Priority: %s | Created: %s
			</div>
			<div>%s</div>
		</div>`, 
			template.HTMLEscapeString(request.Title),
			request.ID,
			template.HTMLEscapeString(request.Priority),
			request.CreatedAt.Format("2006-01-02 15:04:05"),
			template.HTMLEscapeString(request.Description))
	}
}

func handleHTMXForm(w http.ResponseWriter, r *http.Request) {
	// This endpoint could be used to return just the form HTML
	// if you want to reload/reset the form dynamically
	fmt.Fprint(w, `<div>Form reset successfully!</div>`)
}

// Keep the original API endpoints for backward compatibility
func handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var components []string
	if componentsStr, ok := form["affected_components"].(string); ok && componentsStr != "" {
		for _, comp := range strings.Split(componentsStr, ",") {
			comp = strings.TrimSpace(comp)
			if comp != "" {
				components = append(components, comp)
			}
		}
	}

	featureRequest := FeatureRequest{
		ID:                   nextID,
		Title:                getStringValue(form, "title"),
		Description:          getStringValue(form, "description"),
		AcceptanceCriteria:   getStringValue(form, "acceptance_criteria"),
		Priority:             getStringValue(form, "priority"),
		TargetTimeline:       getStringValue(form, "target_timeline"),
		AffectedComponents:   components,
		ExampleUsage:         getStringValue(form, "example_usage"),
		TechnicalConstraints: getStringValue(form, "technical_constraints"),
		CreatedAt:            time.Now(),
		Status:               "submitted",
	}

	if featureRequest.Title == "" || featureRequest.Description == "" ||
		featureRequest.AcceptanceCriteria == "" || featureRequest.Priority == "" {
		writeJSONError(w, "Please fill in all required fields", http.StatusBadRequest)
		return
	}

	featureRequests = append(featureRequests, featureRequest)
	nextID++

	writeJSONResponse(w, APIResponse{
		Success: true,
		Message: "Feature request submitted successfully",
		Data:    featureRequest,
	})
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSONResponse(w, APIResponse{
		Success: true,
		Message: "Feature requests retrieved successfully",
		Data:    featureRequests,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getStringValue(form map[string]interface{}, key string) string {
	if value, ok := form[key].(string); ok {
		return value
	}
	return ""
}

func writeJSONResponse(w http.ResponseWriter, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Message: message,
	})
}
