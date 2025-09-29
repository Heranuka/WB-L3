package render

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type Render struct {
	homeTemplate     *template.Template
	registerTemplate *template.Template
	loginTemplate    *template.Template
	logger           *slog.Logger
}

func New(templatePath string, logger *slog.Logger) *Render {
	return &Render{
		homeTemplate:     template.Must(template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "home.html"))),
		loginTemplate:    template.Must(template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "login.html"))),
		registerTemplate: template.Must(template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "register.html"))),
		logger:           logger,
	}
}

func (r *Render) Home(w http.ResponseWriter, data any) {
	err := r.homeTemplate.Execute(w, data)
	if err != nil {
		r.logger.Error("can not execute home page", slog.String("error", err.Error()))
	}

}

func (r *Render) LoginPage(w http.ResponseWriter) {
	err := r.loginTemplate.Execute(w, nil)
	if err != nil {
		r.logger.Error("can not execute login page", slog.String("error", err.Error()))
	}
}

func (r *Render) RegisterPage(w http.ResponseWriter) {
	err := r.registerTemplate.Execute(w, nil)
	if err != nil {
		r.logger.Error("can not execute register page", slog.String("error", err.Error()))
	}
}
