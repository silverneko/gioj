package main

import (
  "net/http"
  "golang.org/x/net/context"
)

func WelcomeHandler(c context.Context, w http.ResponseWriter, req * http.Request) {
  render("welcome.html", c, w, "")
}

