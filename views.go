package main

import (
  "html/template"
  "log"
  "net/http"
)

var tmpls = make(map[string]*template.Template)

func init() {
  registerTemplate("welcome.html")
  registerTemplate("user/login_form.html")
  registerTemplate("user/register_form.html")
  registerTemplate("user/show.html")
  registerTemplate("user/edit_form.html")
}

func registerTemplate(name string) {
  tmpl, _ := template.ParseFiles("templates/layout.html", "templates/" + name)
  tmpls[name] = tmpl
}

func render(name string, w http.ResponseWriter, r *http.Request, d interface{}, flashes ...string) {
  t, ok := tmpls[name]
  if !ok {
    log.Println("views.go/render() template not found: ", name)
    http.Error(w, "500", 500)
    return
  }
  err := t.Execute(w, map[string]interface{}{
    "Flash": flashes,
    "User": CurrentUser(w, r),
    "Data": d,
  })
  if err != nil {
    log.Println(err)
    http.Error(w, "500: " + err.Error(), 500)
  }
}

