package main

import (
  "html/template"
  "log"
  "net/http"
  "golang.org/x/net/context"
  "github.com/russross/blackfriday"
  "gopkg.in/mgo.v2/bson"
  "github.com/silverneko/gioj/models"
)

var tmpls = make(map[string]*template.Template)

func init() {
  registerTemplate("welcome.html")
  registerTemplate("about.html")

  registerTemplate("user/login_form.html")
  registerTemplate("user/register_form.html")
  registerTemplate("user/show.html")
  registerTemplate("user/edit_form.html")
  registerTemplate("user/index.html")

  registerTemplate("discuss/index.html")

  registerTemplate("problems/index.html")
  registerTemplate("problems/show.html")
  registerTemplate("problems/new.html", "problems/_form.html")
  registerTemplate("problems/edit.html", "problems/_form.html")

  registerTemplate("status/index.html")
  registerTemplate("status/show.html")
  registerTemplate("status/new.html")
  registerTemplate("status/edit.html")
}

func registerTemplate(name... string) {
  tmpl := template.New("layout.html.tmpl").Funcs(template.FuncMap{
    "markdown": func (s string) template.HTML {
      return template.HTML(blackfriday.MarkdownCommon([]byte(s)))
    },
    "truncate": func (s string, l int) string {
      if len(s) > l {
	return s[:l]
      } else {
	return s
      }
    },
    "time": func (id bson.ObjectId) string {
      return id.Time().Format("2006/01/02 15:04:05")
    },
    "humanVerdict": func (verdict int) string {
      switch verdict {
	case models.QUEUED: return "Queued"
	case models.JUDGING: return "Judging"
	case models.AC: return "Accepted"
	case models.WA: return "Wrong Answer"
	case models.RE: return "Runtime Error"
	case models.TLE: return "Time Limit Exceeded"
	case models.MLE: return "Memory Limit Exceeded"
	case models.CE: return "Compilation Error"
	case models.ERR: fallthrough
	default: return "Error"
      }
    },
  })
  tmpl.ParseFiles("templates/layout.html.tmpl")
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
    "User": CurrentUser(c),
    "Data": d,
  })
  if err != nil {
    log.Println(err)
    http.Error(w, "500: " + err.Error(), 500)
  }
}

