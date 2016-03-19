package main

import (
  "html/template"
  "log"
  "net/http"
  "golang.org/x/net/context"
)

var tmpls = make(map[string]*template.Template)

func init() {
  registerTemplate("welcome.html")
  registerTemplate("about.html")
  registerTemplate("user/login_form.html")
  registerTemplate("user/register_form.html")
  registerTemplate("user/show.html")
  registerTemplate("user/edit_form.html")
  registerTemplate("discuss/index.html")
  registerTemplate("problems/index.html")
  registerTemplate("problems/show.html")
  registerTemplate("problems/new.html", "problems/_form.html")
  registerTemplate("problems/edit.html", "problems/_form.html")
}

func registerTemplate(name... string) {
  tmpl, _ := template.ParseFiles("templates/layout.html.tmpl")
  for _, e := range name {
    tmpl.ParseFiles("templates/" + e + ".tmpl")
  }
  tmpls[name[0]] = tmpl
}

func render(name string, c context.Context, w http.ResponseWriter, d interface{}, flashes ...string) {
  t, ok := tmpls[name]
  if !ok {
    log.Println("views.go/render() template not found: ", name)
    http.Error(w, "500", 500)
    return
  }
  err := t.Execute(w, map[string]interface{}{
    "Flash": flashes,
    "User": c.Value("currentUser").(*User),
    "Data": d,
  })
  if err != nil {
    log.Println(err)
    http.Error(w, "500: " + err.Error(), 500)
  }
}

