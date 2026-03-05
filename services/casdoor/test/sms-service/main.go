package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// SMSRequest SMS request structure
type SMSRequest struct {
	Phone       string `json:"phone"`       // Phone number
	PhoneNumber string `json:"phoneNumber"` // Phone number (Casdoor format)
	Code        string `json:"code"`        // Verification code
}

// SMSResponse SMS response structure
type SMSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Mock SMS sending service
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		response := SMSResponse{
			Success: false,
			Message: "Only POST requests are supported",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse request body
	var smsReq SMSRequest

	// First try to parse form data
	r.ParseForm()
	log.Printf("All form parameters: %v", r.Form)

	if len(r.Form) > 0 {
		// Has form data, get parameters from form
		// Casdoor sends parameters named phoneNumber and code
		smsReq.Phone = r.FormValue("phoneNumber") // Modified: get phone number from phoneNumber
		if smsReq.Phone == "" {
			smsReq.Phone = r.FormValue("phone") // Fallback: if phoneNumber is empty, try phone
		}
		smsReq.Code = r.FormValue("code")
		log.Printf("From form data: phone=%s, code=%s", smsReq.Phone, smsReq.Code)
	} else {
		// No form data, try JSON parsing
		err := json.NewDecoder(r.Body).Decode(&smsReq)
		log.Printf("JSON parsing result: err=%v, phone=%s, phoneNumber=%s, code=%s", err, smsReq.Phone, smsReq.PhoneNumber, smsReq.Code)

		// Unify phone number field: use PhoneNumber first, then Phone
		if smsReq.PhoneNumber != "" {
			smsReq.Phone = smsReq.PhoneNumber
		}
	}

	// Log request
	log.Printf("Received SMS sending request - Phone: %s, Code: %s", smsReq.Phone, smsReq.Code)

	// Validate phone number
	if smsReq.Phone == "" {
		response := SMSResponse{
			Success: false,
			Message: "Phone number cannot be empty",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate verification code
	if smsReq.Code == "" {
		response := SMSResponse{
			Success: false,
			Message: "Verification code cannot be empty",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Mock SMS sending process
	log.Printf("Sending verification code to phone number %s: %s", smsReq.Phone, smsReq.Code)

	// Mock network delay
	time.Sleep(100 * time.Millisecond)

	// This is mock sending, in real scenarios it would call actual SMS APIs
	// Such as Alibaba Cloud SMS, Tencent Cloud SMS, etc.

	// Mock successful sending
	response := SMSResponse{
		Success: true,
		Message: fmt.Sprintf("Verification code has been successfully sent to phone number %s", smsReq.Phone),
		Data: map[string]interface{}{
			"phone":     smsReq.Phone,
			"code":      smsReq.Code,
			"timestamp": time.Now().Unix(),
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("SMS sent successfully - Phone: %s", smsReq.Phone)
}

// Health check interface
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "ok",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
		"service": "SMS Verification Code Service",
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Set routes
	http.HandleFunc("/oidc_auth/send/sms", sendSMSHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
		<h1>SMS Verification Code Service</h1>
		<p>Service is running normally</p>
		<p>SMS sending interface: POST /oidc_auth/send/sms</p>
		<p>Health check interface: GET /health</p>
		<p>Current time: %s</p>
		`, time.Now().Format("2006-01-02 15:04:05"))
	})

	port := ":8083"
	log.Printf("SMS verification code service started, listening on port: %s", port)
	log.Println("SMS sending interface: POST http://localhost:8083/oidc_auth/send/sms")
	log.Println("Health check interface: GET http://localhost:8083/health")

	// Start HTTP server
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Service startup failed:", err)
	}
}
