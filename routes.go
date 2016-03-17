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
  mux.HandleFuncC(pat.Get("/register"), RegisterHandler)
  mux.HandleFuncC(pat.Post("/register"), RegisterHandlerP)
  mux.HandleC(pat.Get("/logout"), RequireAuth(LogoutHandler))
  mux.HandleC(pat.Get("/user/:user"), RequireAuth(UserHandler))
  mux.HandleC(pat.Get("/user/:user/edit"), RequireAuth(UserEditHandler))
  mux.HandleC(pat.Post("/user/:user/edit"), RequireAuth(UserEditHandlerP))

  mux.HandleC(pat.Get("/discuss"), RequireAuth(DiscussHandler))
  mux.HandleC(pat.Post("/discuss/new"), RequireAuth(DiscussNewHandler))

  http.Handle("/", mux)
  err := http.ListenAndServe(":4000", http.DefaultServeMux)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

