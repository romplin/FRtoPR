package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

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
    <script src="https://unpkg.com/vue@3/dist/vue.global.js"></script>
    <script src="https://unpkg.com/axios/dist/axios.min.js"></script>
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
                    All Requests
                </button>
            </div>

            <div v-show="activeTab === 'form'">
                <div v-if="message.text" :class="message.type + '-message'">
                    {{ message.text }}
                </div>
                
                <form @submit.prevent="submitForm">
                    <div class="form-group">
                        <label for="title">Feature Title *</label>
                        <input 
                            type="text" 
                            id="title" 
                            v-model="form.title"
                            placeholder="Brief descriptive title for the feature"
                            required
                        >
                    </div>

                    <div class="form-group">
                        <label for="description">Feature Description *</label>
                        <textarea 
                            id="description" 
                            v-model="form.description"
                            placeholder="Detailed description of the feature requirements and functionality"
                            required
                        ></textarea>
                    </div>

                    <div class="form-group">
                        <label for="acceptance_criteria">Acceptance Criteria *</label>
                        <textarea 
                            id="acceptance_criteria" 
                            v-model="form.acceptance_criteria"
                            placeholder="Clear, testable criteria that define when this feature is complete"
                            required
                        ></textarea>
                    </div>

                    <div class="form-group">
                        <label for="priority">Priority Level *</label>
                        <select 
                            id="priority" 
                            v-model="form.priority"
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
                            v-model="form.target_timeline"
                            placeholder="e.g., Next Sprint, Q2 2025, 2 weeks"
                        >
                    </div>

                    <div class="form-group">
                        <label for="affected_components">Affected Components/Modules</label>
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
                        {{ isSubmitting ? 'Submitting...' : 'Submit Feature Request' }}
                    </button>
                    <button type="button" @click="resetForm" class="secondary-btn">Reset Form</button>
                </form>
            </div>

            <div v-show="activeTab === 'list'">
                <div v-if="featureRequests.length === 0" class="feature-item">
                    <div>No feature requests found.</div>
                </div>
                
                <div 
                    v-for="request in featureRequests" 
                    :key="request.id"
                    class="feature-item"
                >
                    <div class="feature-title">{{ request.title }}</div>
                    <div class="feature-meta">
                        ID: {{ request.id }} | 
                        Priority: {{ request.priority }} | 
                        Created: {{ formatDate(request.created_at) }}
                    </div>
                    <div>{{ request.description }}</div>
                </div>
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
                    message: {
                        text: '',
                        type: ''
                    },
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

                async submitForm() {
                    this.isSubmitting = true;
                    this.message = { text: '', type: '' };

                    try {
                        const response = await axios.post('/api/submit', this.form);
                        
                        if (response.data.success) {
                            this.message = {
                                text: 'Feature request submitted successfully! ID: ' + response.data.data.id,
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
                            text: 'Error submitting feature request',
                            type: 'error'
                        };
                    } finally {
                        this.isSubmitting = false;
                        
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
                },

                async loadFeatureRequests() {
                    try {
                        const response = await axios.get('/api/requests');
                        this.featureRequests = response.data.data || [];
                    } catch (error) {
                        console.error('Error loading feature requests:', error);
                    }
                },

                formatDate(dateString) {
                    return new Date(dateString).toLocaleDateString() + ' ' + 
                           new Date(dateString).toLocaleTimeString();
                }
            },

            mounted() {
                this.loadFeatureRequests();
            }
        }).mount('#app');
    </script>
</body>
</html>`
	fmt.Fprint(w, html)
}

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
