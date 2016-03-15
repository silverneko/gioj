package main

import (
  "net/http"
  "goji.io"
  "goji.io/pat"
  "log"
)

func main() {
  mux := goji.NewMux()
  mux.UseC(AuthMiddleware)
  mux.HandleFuncC(pat.Get("/"), WelcomeHandler)

  mux.HandleFuncC(pat.Get("/login"), LoginHandler)
  mux.HandleFuncC(pat.Post("/login"), AuthHandler)
  mux.HandleFuncC(pat.Get("/logout"), LogoutHandler)
  mux.HandleFuncC(pat.Get("/register"), RegisterHandler)
  mux.HandleFuncC(pat.Post("/register"), RegisterHandlerP)
  mux.HandleFuncC(pat.Get("/user/:user"), UserHandler)
  mux.HandleFuncC(pat.Get("/user/:user/edit"), UserEditHandler)
  mux.HandleFuncC(pat.Post("/user/:user/edit"), UserEditHandlerP)

  mux.HandleFuncC(pat.Get("/discuss"), DiscussHandler)
  mux.HandleFuncC(pat.Post("/discuss/new"), DiscussNewHandler)

  http.Handle("/", mux)
  err := http.ListenAndServe(":4000", http.DefaultServeMux)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

