package main

import (
  "net/http"
//  "github.com/gorilla/sessions"
  "log"
)

func WelcomeHandler(w http.ResponseWriter, req * http.Request) {
  session, err := cookieJar.Get(req, AuthSession)
  if err != nil {
    /* Bad cookie */
    log.Println(err)
    session.Options.MaxAge = -1
    session.Save(req, w)
    render("welcome.html", w, "")
  } else {
    username, _ := session.Values["username"].(string)
    render("welcome.html", w, username)
  }
}

