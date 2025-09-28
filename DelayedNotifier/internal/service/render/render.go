package render

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/rs/zerolog"
)

type Renderer interface {
	Home(w http.ResponseWriter)
}

type Render struct {
	homeTemplate *template.Template
	logger       zerolog.Logger
}

func New(templatePath string, logger zerolog.Logger) *Render {
	return &Render{
		homeTemplate: template.Must(template.ParseFiles(fmt.Sprintf("%s/%s", templatePath, "index.html"))),
		logger:       logger,
	}
}

func (r *Render) Home(w http.ResponseWriter) {
	err := r.homeTemplate.Execute(w, nil)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error getting status")
	}

}
