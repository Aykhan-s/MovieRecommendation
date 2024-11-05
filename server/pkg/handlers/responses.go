package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/aykhans/movier/server/pkg/config"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondWithServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func RespondWithJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
		RespondWithServerError(w)
	}
}

func formatNumber(n uint) string {
	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	length := len(s)

	for i, digit := range s {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func RespondWithHTML(w http.ResponseWriter, templateName string, data any, statusCode int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)

	funcMap := template.FuncMap{
		"formatNumber": formatNumber,
	}

	t, err := template.New(templateName).Funcs(funcMap).ParseFiles(config.GetTemplatePath() + "/" + templateName)
	if err != nil {
		log.Printf("error parsing template: %v", err)
		RespondWithServerError(w)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("error executing template: %v", err)
		RespondWithServerError(w)
	}
}
