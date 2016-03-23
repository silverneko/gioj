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
)

// GET /problems
func ProblemsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  db := DBSession{DB.Copy()}
  defer db.Close()
  it := db.C("problems").Find(nil).Sort("_id").Limit(50).Iter()
  var problems []Problem
  if err := it.All(&problems); err != nil {
    log.Println("Problems index: ", err)
  }
  render("problems/index.html", c, w, problems)
}
// GET /problems/:id
func ProblemHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Error(w, "500", 500)
    return
  }
  db := DBSession{DB.Copy()}
  defer db.Close()
  var problem Problem
  if err := db.C("problems").Find(bson.M{"_id": pid}).One(&problem); err != nil {
    http.Error(w, "500", 500)
    return
  }
  render("problems/show.html", c, w, problem)
}
// GET /problems/new
func ProblemNewHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  render("problems/new.html", c, w, "")
}
// POST /problems/new
func ProblemNewHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  user := CurrentUser(c)
  r.ParseMultipartForm(5 * (2 << 20))
  var problem Problem
  decoder.Decode(&problem, r.MultipartForm.Value)
  problem.AuthorName = user.Username
  db := DBSession{DB.Copy()}
  defer db.Close()
  var idx Idx
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
  problem.ID = idx.Seq

  pids := strconv.Itoa(problem.ID)
  file, _, err := r.FormFile("TestdataFile")
  if err == nil {
    path := "./td/" + pids
    os.MkdirAll(path, 0755)
    pathname := path + "/td" + pids + ".zip"
    f, err := os.Create(pathname)
    if err != nil {
      http.Error(w, "500", 500)
      return
    }
    if _, err := io.Copy(f, file); err != nil {
      http.Error(w, "500", 500)
      return
    }
    file.Close()
    f.Close()
    exec.Command("unzip", pathname, "-d"+path).Run()
    problem.Testdata = pathname
  }


  if err := db.C("problems").Insert(problem); err != nil {
    log.Println("Problem create: ", err)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/problems/" + pids, 302)
}
// GET /problems/:id/edit
func ProblemEditHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Error(w, "500", 500)
    return
  }
  var problem Problem
  db := DBSession{DB.Copy()}
  defer db.Close()
  if err := db.C("problems").Find(bson.M{"_id": pid}).One(&problem); err != nil {
    log.Println("Edit problem: ", err)
    http.Error(w, "500", 500)
    return
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
  r.ParseMultipartForm(5 * (2 << 20))
  var problem Problem
  decoder.Decode(&problem, r.MultipartForm.Value)
  problem.ID = pid
  problem.AuthorName = CurrentUser(c).Username

  file, _, err := r.FormFile("TestdataFile")
  if err == nil {
    path := "./td/" + pids
    os.MkdirAll(path, 0755)
    pathname := path + "/td" + pids + ".zip"
    f, err := os.Create(pathname)
    if err != nil {
      http.Error(w, "500", 500)
      return
    }
    if _, err := io.Copy(f, file); err != nil {
      http.Error(w, "500", 500)
      return
    }
    file.Close()
    f.Close()
    exec.Command("unzip", pathname, "-d"+path).Run()
    problem.Testdata = pathname
  }

  db := DBSession{DB.Copy()}
  defer db.Close()
  if _, err := db.C("problems").Upsert(bson.M{"_id": pid}, bson.M{"$set": problem}); err != nil {
    log.Println("Problem update: ", err, pid)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/problems/" + pids, 302)
}

