package service

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type ServiceRender interface {
	Home(w http.ResponseWriter)
}
type Render struct {
	homeTemplate  *template.Template
	serviceRender ServiceRender
	logger        *slog.Logger
}

func NewRender(templatePath string, serviceSender ServiceRender, logger *slog.Logger) *Render {
	return &Render{
		homeTemplate:  template.Must(template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "home.html"))),
		serviceRender: serviceSender,
		logger:        logger,
	}
}

func (r *Render) Home(w http.ResponseWriter) {
	err := r.homeTemplate.Execute(w, nil)
	if err != nil {
		r.logger.Error("can not execute home page", slog.String("error", err.Error()))
	}

}
