package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/rs/zerolog"
)

type Renderer interface {
	Home(w http.ResponseWriter, data any)
	LoginPage(w http.ResponseWriter)
	RegisterPage(w http.ResponseWriter)
}

type Render struct {
	homeTemplate     *template.Template
	registerTemplate *template.Template
	loginTemplate    *template.Template
	logger           zerolog.Logger
}

func New(templatePath string, logger zerolog.Logger) *Render {
	homeTpl, err := template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "home.html"))
	if err != nil {
		log.Fatalf("failed to parse home.html template: %v", err)
	}

	loginTpl, err := template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "login.html"))
	if err != nil {
		log.Fatalf("failed to parse login.html template: %v", err)
	}

	registerTpl, err := template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "register.html"))
	if err != nil {
		log.Fatalf("failed to parse register.html template: %v", err)
	}

	return &Render{
		homeTemplate:     homeTpl,
		loginTemplate:    loginTpl,
		registerTemplate: registerTpl,
		logger:           logger,
	}
}

func (r *Render) Home(w http.ResponseWriter, data any) {
	var buf bytes.Buffer
	if err := r.homeTemplate.Execute(&buf, data); err != nil {
		r.logger.Error().Err(err).Msg("template execute error in Home")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, err := buf.WriteTo(w)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to write template output in Home")
	}

}

func (r *Render) LoginPage(w http.ResponseWriter) {
	if err := r.loginTemplate.Execute(w, nil); err != nil {
		r.logger.Error().Err(err).Msg("Failed to execute login page template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (r *Render) RegisterPage(w http.ResponseWriter) {
	if err := r.registerTemplate.Execute(w, nil); err != nil {
		r.logger.Error().Err(err).Msg("Failed to execute register page template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
