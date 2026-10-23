package controller

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"chat/common"
	"chat/database"
)

func Top(w http.ResponseWriter, r *http.Request) {
	cfg := common.LoadConfig()

	var session database.SessionStruct
	var err error
	var tmplPath string
	var domain string
	var googleClientID string
	var sessionJSON template.JS

	switch {
	case strings.Contains(r.URL.Path, "/sign/"):
		tmplPath = "view/signGoogle.html"
		domain = cfg.Domain
		googleClientID = cfg.GoogleClientID

		session, _ = common.SessionGet(w, r)

		session, err = common.PushReGenerate(session)
		if err != nil {
			log.Printf(
				"ReGenerateCSRF failed: %v; path=%s",
				err,
				r.URL.Path,
			)
		}

		b, err := json.Marshal(session)
		if err != nil {
			log.Printf(
				"session JSON encoding failed: %v; path=%s",
				err,
				r.URL.Path,
			)
		} else {
			sessionJSON = template.JS(b)
		}

	case strings.Contains(r.URL.Path, "/pushSubscription/"):
		tmplPath = "view/pushSubscription.html"

		session, err = common.SessionGet(w, r)
		if err != nil {
			log.Printf(
				"SessionGet failed: %v; path=%s",
				err,
				r.URL.Path,
			)
			http.Error(
				w,
				"Error SessionGet",
				http.StatusInternalServerError,
			)
			return
		}

	default:
		tmplPath = "view/index.html"
	}

	type View struct {
		Session        database.SessionStruct
		Domain         string
		GoogleClientID string
		CacheV         string
		SessionJS      template.JS
	}

	view := View{
		Session:        session,
		Domain:         domain,
		GoogleClientID: googleClientID,
		CacheV:         cfg.CacheV,
		SessionJS:      sessionJSON,
	}

	tpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf(
			"Template parsing failed: %v; path=%s",
			err,
			tmplPath,
		)
		http.Error(
			w,
			"Error loading template",
			http.StatusInternalServerError,
		)
		return
	}

	if err := tpl.Execute(w, view); err != nil {
		log.Printf(
			"Template execution failed: %v; path=%s",
			err,
			r.URL.Path,
		)
		http.Error(
			w,
			"Error rendering template",
			http.StatusInternalServerError,
		)
		return
	}
}