package main

import (
  "net/http"
  "github.com/gorilla/context"
  "goji.io"
  "goji.io/pat"
  "log"
)

func main() {
  mux := goji.NewMux()
  mux.HandleFunc(pat.Get("/"), WelcomeHandler)
  mux.HandleFunc(pat.Get("/login"), LoginHandler)
  mux.HandleFunc(pat.Post("/login"), AuthHandler)
  mux.HandleFunc(pat.Get("/logout"), LogoutHandler)
  mux.HandleFunc(pat.Get("/register"), RegisterHandler)
  mux.HandleFunc(pat.Post("/register"), RegisterHandlerP)
  mux.HandleFuncC(pat.Get("/user/:user"), UserHandler)
  mux.HandleFuncC(pat.Get("/user/:user/edit"), UserEditHandler)
  mux.HandleFuncC(pat.Post("/user/:user/edit"), UserEditHandlerP)

  http.Handle("/", mux)
  err := http.ListenAndServe(":4000", context.ClearHandler(http.DefaultServeMux))
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

