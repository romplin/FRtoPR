package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// FeatureRequest represents the structure of a feature request
type FeatureRequest struct {
	ID                  int       `json:"id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	AcceptanceCriteria  string    `json:"acceptance_criteria"`
	Priority            string    `json:"priority"`
	TargetTimeline      string    `json:"target_timeline"`
	AffectedComponents  []string  `json:"affected_components"`
	ExampleUsage        string    `json:"example_usage"`
	TechnicalConstraints string   `json:"technical_constraints"`
	CreatedAt           time.Time `json:"created_at"`
	Status              string    `json:"status"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// In-memory storage (replace with database in production)
var featureRequests []FeatureRequest
var nextID = 1

// GitHub MCP Server configuration
type GitHubConfig struct {
	ServerURL string
	Token     string
}

var githubConfig GitHubConfig

func init() {
	githubConfig = GitHubConfig{
		ServerURL: os.Getenv("GITHUB_MCP_SERVER_URL"),
		Token:     os.Getenv("GITHUB_MCP_TOKEN"),
	}
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Feature Request System</title>
    <script src="https://unpkg.com/vue@3/dist/vue.global.js"></script>
    <script src="https://unpkg.com/axios/dist/axios.min.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
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
        .priority-select {
            background-color: #fff;
        }
        .components-help {
            font-size: 12px;
            color: #666;
            margin-bottom: 10px;
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
        .secondary-btn:hover {
            background-color: #545b62;
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
        .feature-list {
            margin-top: 20px;
        }
        .feature-item {
            background: #f8f9fa;
            padding: 20px;
            margin-bottom: 15px;
            border-radius: 4px;
            border-left: 4px solid #007bff;
        }
        .feature-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .feature-title {
            font-size: 18px;
            font-weight: 600;
            color: #333;
        }
        .feature-priority {
            padding: 4px 8px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }
        .priority-high { background-color: #ff4444; color: white; }
        .priority-medium { background-color: #ffaa00; color: white; }
        .priority-low { background-color: #00aa00; color: white; }
        .feature-meta {
            font-size: 12px;
            color: #666;
            margin-bottom: 10px;
        }
        .feature-description {
            margin-bottom: 15px;
            line-height: 1.5;
        }
        .feature-details {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-top: 15px;
        }
        .detail-section {
            background: white;
            padding: 15px;
            border-radius: 4px;
            border: 1px solid #e9ecef;
        }
        .detail-title {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
        }
        .components-list {
            display: flex;
            flex-wrap: wrap;
            gap: 5px;
        }
        .component-tag {
            background-color: #e9ecef;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 12px;
            color: #495057;
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
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        .loading {
            text-align: center;
            padding: 20px;
            color: #666;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #007bff;
            border-radius: 50%;
            width: 20px;
            height: 20px;
            animation: spin 1s linear infinite;
            display: inline-block;
            margin-right: 10px;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        .validation-error {
            color: #721c24;
            font-size: 12px;
            margin-top: 5px;
        }
        .field-error {
            border-color: #dc3545;
        }
        .fade-enter-active, .fade-leave-active {
            transition: opacity 0.3s;
        }
        .fade-enter-from, .fade-leave-to {
            opacity: 0;
        }
        .list-enter-active, .list-leave-active {
            transition: all 0.3s ease;
        }
        .list-enter-from, .list-leave-to {
            opacity: 0;
            transform: translateY(-10px);
        }
    </style>
</head>
<body>
    <div id="app">
        <div class="container">
            <h1>Feature Request System</h1>
            
            <div class="tabs">
                <button 
                    class="tab" 
                    :class="{ active: activeTab === 'form' }"
                    @click="activeTab = 'form'"
                >
                    New Request
                </button>
                <button 
                    class="tab" 
                    :class="{ active: activeTab === 'list' }"
                    @click="setActiveTab('list')"
                >
                    All Requests ({{ featureRequests.length }})
                </button>
            </div>

            <!-- Form Tab -->
            <div v-show="activeTab === 'form'" class="tab-content active">
                <transition name="fade">
                    <div v-if="message.text" :class="message.type + '-message'">
                        {{ message.text }}
                    </div>
                </transition>
                
                <form @submit.prevent="submitForm">
                    <div class="form-group">
                        <label for="title">Feature Title *</label>
                        <input 
                            type="text" 
                            id="title" 
                            v-model="form.title"
                            :class="{ 'field-error': errors.title }"
                            placeholder="Brief descriptive title for the feature"
                            required
                        >
                        <div v-if="errors.title" class="validation-error">{{ errors.title }}</div>
                    </div>

                    <div class="form-group">
                        <label for="description">Feature Description *</label>
                        <textarea 
                            id="description" 
                            v-model="form.description"
                            :class="{ 'field-error': errors.description }"
                            placeholder="Detailed description of the feature requirements and functionality"
                            required
                        ></textarea>
                        <div v-if="errors.description" class="validation-error">{{ errors.description }}</div>
                    </div>

                    <div class="form-group">
                        <label for="acceptance_criteria">Acceptance Criteria *</label>
                        <textarea 
                            id="acceptance_criteria" 
                            v-model="form.acceptance_criteria"
                            :class="{ 'field-error': errors.acceptance_criteria }"
                            placeholder="Clear, testable criteria that define when this feature is complete"
                            required
                        ></textarea>
                        <div v-if="errors.acceptance_criteria" class="validation-error">{{ errors.acceptance_criteria }}</div>
                    </div>

                    <div class="form-group">
                        <label for="priority">Priority Level *</label>
                        <select 
                            id="priority" 
                            v-model="form.priority"
                            :class="{ 'field-error': errors.priority }"
                            class="priority-select"
                            required
                        >
                            <option value="">Select priority</option>
                            <option value="high">High - Critical/Urgent</option>
                            <option value="medium">Medium - Important</option>
                            <option value="low">Low - Nice to have</option>
                        </select>
                        <div v-if="errors.priority" class="validation-error">{{ errors.priority }}</div>
                    </div>

                    <div class="form-group">
                        <label for="target_timeline">Target Timeline</label>
                        <input 
                            type="text" 
                            id="target_timeline" 
                            v-model="form.target_timeline"
                            placeholder="e.g., Next Sprint, Q2 2025, 2 weeks"
                        >
                    </div>

                    <div class="form-group">
                        <label for="affected_components">Affected Components/Modules</label>
                        <div class="components-help">Enter components separated by commas (e.g., "authentication, user-profile, notifications")</div>
                        <input 
                            type="text" 
                            id="affected_components" 
                            v-model="form.affected_components"
                            placeholder="api, frontend, database, auth-service"
                        >
                    </div>

                    <div class="form-group">
                        <label for="example_usage">Example Usage Scenarios</label>
                        <textarea 
                            id="example_usage" 
                            v-model="form.example_usage"
                            placeholder="Provide specific examples of how users would interact with this feature"
                        ></textarea>
                    </div>

                    <div class="form-group">
                        <label for="technical_constraints">Technical Constraints or Preferences</label>
                        <textarea 
                            id="technical_constraints" 
                            v-model="form.technical_constraints"
                            placeholder="Any technical limitations, preferred technologies, or implementation constraints"
                        ></textarea>
                    </div>

                    <button type="submit" :disabled="isSubmitting">
                        <span v-if="isSubmitting" class="spinner"></span>
                        {{ isSubmitting ? 'Submitting...' : 'Submit Feature Request' }}
                    </button>
                    <button type="button" @click="resetForm" class="secondary-btn">Reset Form</button>
                </form>
            </div>

            <!-- List Tab -->
            <div v-show="activeTab === 'list'" class="tab-content">
                <div v-if="isLoading" class="loading">
                    <div class="spinner"></div>
                    Loading feature requests...
                </div>
                
                <div v-else-if="featureRequests.length === 0" class="feature-item">
                    <div class="feature-description">No feature requests found.</div>
                </div>
                
                <transition-group name="list" tag="div" class="feature-list">
                    <div 
                        v-for="request in featureRequests" 
                        :key="request.id"
                        class="feature-item"
                    >
                        <div class="feature-header">
                            <div class="feature-title">{{ request.title }}</div>
                            <div class="feature-priority" :class="'priority-' + request.priority">
                                {{ request.priority }}
                            </div>
                        </div>
                        <div class="feature-meta">
                            ID: {{ request.id }} | 
                            Created: {{ formatDate(request.created_at) }} | 
                            Status: {{ request.status }}
                        </div>
                        <div class="feature-description">{{ request.description }}</div>
                        <div class="feature-details">
                            <div class="detail-section">
                                <div class="detail-title">Acceptance Criteria</div>
                                <div>{{ request.acceptance_criteria }}</div>
                            </div>
                            <div class="detail-section">
                                <div class="detail-title">Target Timeline</div>
                                <div>{{ request.target_timeline || 'Not specified' }}</div>
                            </div>
                            <div class="detail-section">
                                <div class="detail-title">Affected Components</div>
                                <div class="components-list">
                                    <span 
                                        v-for="component in request.affected_components" 
                                        :key="component"
                                        class="component-tag"
                                    >
                                        {{ component }}
                                    </span>
                                    <span v-if="!request.affected_components || request.affected_components.length === 0">
                                        None specified
                                    </span>
                                </div>
                            </div>
                            <div class="detail-section">
                                <div class="detail-title">Example Usage</div>
                                <div>{{ request.example_usage || 'Not provided' }}</div>
                            </div>
                            <div class="detail-section">
                                <div class="detail-title">Technical Constraints</div>
                                <div>{{ request.technical_constraints || 'None specified' }}</div>
                            </div>
                        </div>
                    </div>
                </transition-group>
            </div>
        </div>
    </div>

    <script>
        const { createApp } = Vue;

        createApp({
            data() {
                return {
                    activeTab: 'form',
                    isSubmitting: false,
                    isLoading: false,
                    message: {
                        text: '',
                        type: ''
                    },
                    errors: {},
                    form: {
                        title: '',
                        description: '',
                        acceptance_criteria: '',
                        priority: '',
                        target_timeline: '',
                        affected_components: '',
                        example_usage: '',
                        technical_constraints: ''
                    },
                    featureRequests: []
                }
            },
            methods: {
                setActiveTab(tab) {
                    this.activeTab = tab;
                    if (tab === 'list') {
                        this.loadFeatureRequests();
                    }
                },
                
                validateForm() {
                    this.errors = {};
                    let isValid = true;

                    if (!this.form.title.trim()) {
                        this.errors.title = 'Title is required';
                        isValid = false;
                    }

                    if (!this.form.description.trim()) {
                        this.errors.description = 'Description is required';
                        isValid = false;
                    }

                    if (!this.form.acceptance_criteria.trim()) {
                        this.errors.acceptance_criteria = 'Acceptance criteria is required';
                        isValid = false;
                    }

                    if (!this.form.priority) {
                        this.errors.priority = 'Priority is required';
                        isValid = false;
                    }

                    return isValid;
                },

                async submitForm() {
                    if (!this.validateForm()) {
                        return;
                    }

                    this.isSubmitting = true;
                    this.message = { text: '', type: '' };

                    try {
                        const response = await axios.post('/api/submit', this.form);
                        
                        if (response.data.success) {
                            this.message = {
                                text: `Feature request submitted successfully! ID: ' + response.data.data.id,
                                type: 'success'
                            };
                            this.resetForm();
                        } else {
                            this.message = {
                                text: response.data.message || 'Error submitting feature request',
                                type: 'error'
                            };
                        }
                    } catch (error) {
                        this.message = {
                            text: error.response?.data?.message || 'Error submitting feature request',
                            type: 'error'
                        };
                    } finally {
                        this.isSubmitting = false;
                        
                        // Clear message after 5 seconds
                        setTimeout(() => {
                            this.message = { text: '', type: '' };
                        }, 5000);
                    }
                },

                resetForm() {
                    this.form = {
                        title: '',
                        description: '',
                        acceptance_criteria: '',
                        priority: '',
                        target_timeline: '',
                        affected_components: '',
                        example_usage: '',
                        technical_constraints: ''
                    };
                    this.errors = {};
                },

                async loadFeatureRequests() {
                    this.isLoading = true;
                    try {
                        const response = await axios.get('/api/requests');
                        this.featureRequests = response.data.data || [];
                    } catch (error) {
                        console.error('Error loading feature requests:', error);
                        this.message = {
                            text: 'Error loading feature requests',
                            type: 'error'
                        };
                    } finally {
                        this.isLoading = false;
                    }
                },

                formatDate(dateString) {
                    return new Date(dateString).toLocaleDateString() + ' ' + 
                           new Date(dateString).toLocaleTimeString();
                }
            },

            mounted() {
                // Load feature requests on initial load
                this.loadFeatureRequests();
            }
        }).mount('#app');
    </script>
</body>
</html>
`

func main() {
	// Serve static files (for the frontend)
	http.HandleFunc("/", handleHome)
	
	// API endpoints
	http.HandleFunc("/api/submit", handleAPISubmit)
	http.HandleFunc("/api/requests", handleAPIRequests)
	http.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlTemplate)
}

func handleAPISubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse affected components
	var components []string
	if componentsStr, ok := form["affected_components"].(string); ok && componentsStr != "" {
		for _, comp := range strings.Split(componentsStr, ",") {
			comp = strings.TrimSpace(comp)
			if comp != "" {
				components = append(components, comp)
			}
		}
	}

	// Create feature request
	featureRequest := FeatureRequest{
		ID:                  nextID,
		Title:               getStringValue(form, "title"),
		Description:         getStringValue(form, "description"),
		AcceptanceCriteria:  getStringValue(form, "acceptance_criteria"),
		Priority:            getStringValue(form, "priority"),
		TargetTimeline:      getStringValue(form, "target_timeline"),
		AffectedComponents:  components,
		ExampleUsage:        getStringValue(form, "example_usage"),
		TechnicalConstraints: getStringValue(form, "technical_constraints"),
		CreatedAt:           time.Now(),
		Status:              "submitted",
	}

	// Validate required fields
	if featureRequest.Title == "" || featureRequest.Description == "" || 
		featureRequest.AcceptanceCriteria == "" || featureRequest.Priority == "" {
		writeJSONError(w, "Please fill in all required fields", http.StatusBadRequest)
		return
	}

	// Save to in-memory storage
	featureRequests = append(featureRequests, featureRequest)
	nextID++

	// Submit to GitHub MCP Server
	if err := submitToGitHubMCP(featureRequest); err != nil {
		log.Printf("Error submitting to GitHub MCP: %v", err)
		writeJSONError(w, fmt.Sprintf("Feature request saved but failed to submit to GitHub: %v", err), http.StatusInternalServerError)
		return
	}

	// Success response
	writeJSONResponse(w, APIResponse{
		Success: true,
		Message: "Feature request submitted successfully",
		Data:    featureRequest,
	})
}

func handleAPIRequests(w http.ResponseWriter, r *http.Request) {
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

// Helper functions
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

// submitToGitHubMCP submits the feature request to GitHub MCP server
func submitToGitHubMCP(fr FeatureRequest) error {
	if githubConfig.ServerURL == "" {
		return fmt.Errorf("GitHub MCP server URL not configured")
	}

	// Prepare the payload for GitHub MCP
	payload := map[string]interface{}{
		"title": fr.Title,
		"body": fmt.Sprintf(`## Feature Description
%s

## Acceptance Criteria
%s

## Priority
%s

## Target Timeline
%s

## Affected Components
%s

## Example Usage
%s

## Technical Constraints
%s

---
*Created: %s*
*ID: %d*`,
			fr.Description,
			fr.AcceptanceCriteria,
			fr.Priority,
			fr.TargetTimeline,
			strings.Join(fr.AffectedComponents, ", "),
			fr.ExampleUsage,
			fr.TechnicalConstraints,
			fr.CreatedAt.Format("2006-01-02 15:04:05"),
			fr.ID),
		"labels": []string{"feature-request", "priority-" + fr.Priority},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create HTTP request to GitHub MCP server
	req, err := http.NewRequest("POST", githubConfig.ServerURL+"/api/issues", strings.NewReader(string(jsonPayload)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if githubConfig.Token != "" {
		req.Header.Set("Authorization", "Bearer "+githubConfig.Token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit to GitHub MCP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("GitHub MCP server returned status: %d", resp.StatusCode)
	}

	return nil
}

