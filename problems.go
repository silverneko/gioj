package main

import (
  "net/http"
  "goji.io/pat"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/net/context"
  "log"
  "strconv"
  "os"
  "os/exec"
  "io"
  "github.com/silverneko/gioj/models"
)

// GET /problems
func ProblemsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  page, err := strconv.Atoi(r.FormValue("page"))
  if err != nil || page < 1 {
    page = 1
  }
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  var idx models.Idx
  if err := db.C("counters").Find(bson.M{"_id": "problems"}).One(&idx); err != nil {
    log.Println("Counter retrieve (problem): ", err)
  }
  var pages []int
  for i, n := 1, (idx.Seq + 49) / 50; i <= n; i++ {
    pages = append(pages, i)
  }

  it := db.C("problems").Find(nil).Sort("_id").Skip((page-1)*50).Limit(50).Iter()
  var problems []models.Problem
  if err := it.All(&problems); err != nil {
    log.Println("Problems index: ", err)
  }
  render("problems/index.html", c, w, struct{
    Problems []models.Problem
    Page int
    Pages []int
  }{
    problems,
    page,
    pages,
  })
}
// GET /problems/:id
func ProblemHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Redirect(w, r, "/404", 303)
    return
  }
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  var problem models.Problem
  if err := db.C("problems").Find(bson.M{"_id": pid}).One(&problem); err != nil {
    http.Redirect(w, r, "/404", 303)
    return
  }
  render("problems/show.html", c, w, problem)
}
// GET /problems/new
func ProblemNewHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  render("problems/new.html", c, w, new(models.Problem))
}
// POST /problems/new
func ProblemNewHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  var idx models.Idx
  _, err := db.C("counters").Find(bson.M{"_id": "problems"}).Apply(mgo.Change{
    Update: bson.M{"$inc": bson.M{"seq": 1}},
    Upsert: true,
    ReturnNew: true,
  }, &idx)
  if err != nil {
    log.Println("Counter retrieve (problem): ", err)
    http.Error(w, "500", 500)
    return
  }
  var problem models.Problem

  problem.ID = idx.Seq
  if err := parseProblemForm(c, r, &problem); err != nil {
    http.Error(w, "500", 500)
    return
  }

  if err := db.C("problems").Insert(problem); err != nil {
    log.Println("Problem create: ", err)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/problems/" + strconv.Itoa(problem.ID), 302)
}
// GET /problems/:id/edit
func ProblemEditHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Redirect(w, r, "/404", 303)
    return
  }
  var problem models.Problem
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  if err := db.C("problems").Find(bson.M{"_id": pid}).One(&problem); err != nil {
    log.Println("Edit problem: ", err)
    problem.ID = pid
  }
  render("problems/edit.html", c, w, problem)
}
// POST /problems/:id/edit
func ProblemEditHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  pids := pat.Param(c, "id")
  pid, err := strconv.Atoi(pids)
  if err != nil {
    http.Error(w, "500", 500)
    return
  }

  var problem models.Problem
  problem.ID = pid
  if err := parseProblemForm(c, r, &problem); err != nil {
    http.Error(w, "500", 500)
    return
  }

  db := models.DBSession{DB.Copy()}
  defer db.Close()
  if _, err := db.C("problems").Upsert(bson.M{"_id": pid}, bson.M{"$set": problem}); err != nil {
    log.Println("Problem update: ", err, pid)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/problems/" + pids, 302)
}

func parseProblemForm(c context.Context, r *http.Request, problem *models.Problem) error {
  pids := strconv.Itoa(problem.ID)
  r.ParseMultipartForm(5 * (2 << 20))
  decoder.Decode(problem, r.MultipartForm.Value)
  problem.AuthorName = CurrentUser(c).Username
  if file, _, err := r.FormFile("TestdataFile"); err == nil {
    defer file.Close()
    path := "./td/" + pids
    os.RemoveAll(path)
    os.MkdirAll(path, 0755)
    pathname := path + "/td" + pids + ".zip"
    f, err := os.Create(pathname)
    if err != nil {
      return err
    }
    defer f.Close()
    if _, err := io.Copy(f, file); err != nil {
      return err
    }
    exec.Command("unzip", pathname, "-d"+path).Run()
    problem.Testdata = pathname
  }
  return nil
}

