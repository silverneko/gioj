package main

import (
  "html/template"
  "log"
  "net/http"
)

var tmpls = make(map[string]*template.Template)

func init() {
  registerTemplate("welcome.html")
  registerTemplate("login_form.html")
  registerTemplate("register_form.html")
}

func registerTemplate(name string) {
  tmpls[name], _ = template.ParseFiles("templates/layout.html", "templates/" + name)
}

func render(name string, w http.ResponseWriter, d interface{}) {
  t, ok := tmpls[name]
  if !ok {
    log.Println("views.go/render() template not found: ", name)
    http.Error(w, "500", 500)
    return
  }
  err := t.Execute(w, d)
  if err != nil {
    log.Println(err)
    http.Error(w, "500", 500)
  }
}

