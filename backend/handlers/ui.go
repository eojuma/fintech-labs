package handlers

import (
    "html/template"
    "log"
    "net/http"
    "fintech-labs/services"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Get the username from the session cookie (already authenticated by middleware)
    cookie, _ := r.Cookie("session_user")
    username := cookie.Value

    // 2. Fetch the account and the User data from the database
    account, err := services.GetBalanceProcess(username)
    if err != nil {
        log.Printf("Dashboard Error for %s: %v", username, err)
        http.Error(w, "Account not found. Please register first.", http.StatusNotFound)
        return
    }

    // 3. Find and Parse the HTML Template
    // Note: The path "templates/dashboard.html" is relative to where you run 'go run .'
    tmpl, err := template.ParseFiles("templates/dashboard.html")
    if err != nil {
        // This log will appear in your terminal to help you debug the path
        log.Printf("CRITICAL TEMPLATE ERROR: Could not find dashboard.html. Error: %v", err)
        
        // Sending a 404 to the browser because the "Page" (template) wasn't found
        http.Error(w, "Internal Server Error: UI Template not found", http.StatusNotFound)
        return
    }

    // 4. Send the data to the browser
    err = tmpl.Execute(w, account)
    if err != nil {
        log.Printf("EXECUTION ERROR: Failed to render template for %s: %v", username, err)
        http.Error(w, "Error rendering the page", http.StatusInternalServerError)
    }
}