package main

import (
  "net/http"
)

func WelcomeHandler(w http.ResponseWriter, req * http.Request) {
  render("welcome.html", w, req, "")
}

