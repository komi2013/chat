package controller

import (
	"context"
	"html/template"
  "log"
  "net/http"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/common"
  "chat/collection"
)

func Top(w http.ResponseWriter, r *http.Request) {
	var session collection.SessionStruct
	var tmplPath string
	switch {
	case strings.Contains(r.URL.Path, "/sign/"):
		tmplPath = "view/signTmp.tmpl"
		// tmplPath = "view/signGoogle.tmpl"
	case strings.Contains(r.URL.Path, "/pushSubscription/"):
		tmplPath = "view/pushSubscription.tmpl"
	  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	  defer cancel()
	  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	  if err != nil {
	    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
	  }
	  defer c.Disconnect(ctx)
	  db1 := c.Database(common.MongoDb1)
		session, err = common.SessionGet(db1, w, r)
		if err != nil {
			log.Printf("SessionGet: %v; Req: ", err, r.URL.Path, r.Form)
			http.Error(w, "Error SessionGet", http.StatusInternalServerError)
			return
		}
		session, err = common.ReGenerateCSRF(db1, session)
		if err != nil {
			log.Printf("ReGenerateCSRF: %v; Req: ", err, r.URL.Path, r.Form)
			http.Error(w, "Error ReGenerateCSRF", http.StatusInternalServerError)
			return
		}
	case r.URL.Path == "/setting/":
		tmplPath = "public/index.html"
	  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	  defer cancel()
	  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	  if err != nil {
	    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
	  }
	  defer c.Disconnect(ctx)
	  db1 := c.Database(common.MongoDb1)
		session, err = common.SessionGet(db1, w, r)
		if err != nil {
			log.Printf("SessionGet: %v; Req: ", err, r.URL.Path, r.Form)
			http.Error(w, "Error SessionGet", http.StatusInternalServerError)
			return
		}
		session, err = common.ReGenerateData(db1, session)
		if err != nil {
			log.Printf("ReGenerateCSRF: %v; Req: ", err, r.URL.Path, r.Form)
			http.Error(w, "Error ReGenerateCSRF", http.StatusInternalServerError)
			return
		}
	default:
		tmplPath = "public/index.html"
	}
	type View struct {
		Session collection.SessionStruct
	}
	var view View
	view.Session = session
	tpl := template.Must(template.ParseFiles(tmplPath))
	if err := tpl.Execute(w, view); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Template execution: %v; Req: ", err, r.URL.Path, r.Form)
		return
	}
}
