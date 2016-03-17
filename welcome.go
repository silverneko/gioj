package main

import (
  "net/http"
  "golang.org/x/net/context"
)

func WelcomeHandler(c context.Context, w http.ResponseWriter, r * http.Request) {
  render("welcome.html", c, w, "")
}

func AboutHandler(c context.Context, w http.ResponseWriter, r * http.Request) {
  render("about.html", c, w, "")
}

