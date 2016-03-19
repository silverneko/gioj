package main

import (
  "net/http"
  "goji.io/pat"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/net/context"
  "log"
  "strconv"
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
  user := c.Value("currentUser").(*User)
  var problem Problem
  Decode(&problem, r)
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
  err = db.C("problems").Insert(bson.M{
    "_id": problem.ID,
    "name": problem.Name,
    "content": problem.Content,
    "authorname": problem.AuthorName,
  })
  if err != nil {
    log.Println("Problem create: ", err)
    http.Error(w, "500", 500)
    return
  }
  pid := strconv.Itoa(problem.ID)
  http.Redirect(w, r, "/problems/" + pid, 302)
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
  var problem Problem
  Decode(&problem, r)
  problem.ID = pid
  db := DBSession{DB.Copy()}
  defer db.Close()
  if _, err := db.C("problems").Upsert(bson.M{"_id": pid}, bson.M{"$set": problem}); err != nil {
    log.Println("Problem update: ", err, pid)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/problems/" + pids, 302)
}

