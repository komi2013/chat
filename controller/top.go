package controller

import (
	// "context"
	// "errors"
	"encoding/json"
	"html/template"
  "log"
  "net/http"
  "strings"
  // "time"

  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/common"
  "chat/collection"
)

func Top(w http.ResponseWriter, r *http.Request) {
  cfg := common.LoadConfig()

	var session collection.SessionStruct
	var err error
	var tmplPath string
	var domain string
	var googleClientID string
	var sessionJSON template.JS
	switch {
	case strings.Contains(r.URL.Path, "/sign/"):
		// tmplPath = "view/signTmp.tmpl"
		tmplPath = "view/signGoogle.html"
		domain = cfg.Domain
		googleClientID = cfg.GoogleClientID

		session, _ = common.SessionGet(w, r)
		session, err = common.PushReGenerate(session)
		if err != nil {
			log.Printf("ReGenerateCSRF: %v; Req: ", err, r.URL.Path, r.Form)
		}
		b, err := json.Marshal(session)
		if err != nil {
			log.Printf("sessionJSON: %v; Req: ", err, r.URL.Path, r.Form)
		}
		sessionJSON = template.JS(b)
	case strings.Contains(r.URL.Path, "/pushSubscription/"):
		tmplPath = "view/pushSubscription.html"
		session, err = common.SessionGet(w, r)
		if err != nil {
			log.Printf("SessionGet: %v; Req: ", err, r.URL.Path, r.Form)
			http.Error(w, "Error SessionGet", http.StatusInternalServerError)
			return
		}
		// session, err = common.PushReGenerate(session)
		// if err != nil {
		// 	log.Printf("ReGenerateCSRF: %v; Req: ", err, r.URL.Path, r.Form)
		// 	http.Error(w, "Error ReGenerateCSRF", http.StatusInternalServerError)
		// 	return
		// }
	default:
		tmplPath = "view/index.html"
	}
	type View struct {
		Session collection.SessionStruct
		Domain string
		GoogleClientID string
		CacheV string
		SessionJS template.JS
	}
	var view View
	view.Session = session
	view.Domain = domain
	view.GoogleClientID = googleClientID
	view.CacheV = cfg.CacheV
	view.SessionJS = sessionJSON

	tpl := template.Must(template.ParseFiles(tmplPath))
	if err := tpl.Execute(w, view); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template execution: %v; Req: ", err, r.URL.Path, r.Form)
		return
	}
}
