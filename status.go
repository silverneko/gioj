package main

import (
  "net/http"
  "goji.io/pat"
  _ "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/net/context"
  "log"
  "strconv"
  "github.com/silverneko/gioj/models"
)

// GET /status
func StatusHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  it := db.C("submissions").Find(nil).Sort("-_id").Limit(50).Iter()
  var submissions []models.Submission
  if err := it.All(&submissions); err != nil {
    log.Println("Status index: ", err)
  }
  render("status/index.html", c, w, submissions)
}

// GET /status/:id
func StatusShowHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  sidh := pat.Param(c, "id")
  if !bson.IsObjectIdHex(sidh) {
    http.Error(w, "500", 500)
    return
  }
  sid := bson.ObjectIdHex(sidh)
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  var submission models.Submission
  if err := db.C("submissions").Find(bson.M{"_id": sid}).One(&submission); err != nil {
    http.Error(w, "500", 500)
    return
  }
  render("status/show.html", c, w, submission)
}

// GET /status/:id/edit
func StatusEditHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  sidh := pat.Param(c, "id")
  if !bson.IsObjectIdHex(sidh) {
    http.Error(w, "500", 500)
    return
  }
  sid := bson.ObjectIdHex(sidh)
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  var submission models.Submission
  if err := db.C("submissions").Find(bson.M{"_id": sid}).One(&submission); err != nil {
    http.Error(w, "500", 500)
    return
  }
  render("status/edit.html", c, w, submission)
}

// POST /status/:id/edit
func StatusEditHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  sidh := pat.Param(c, "id")
  if !bson.IsObjectIdHex(sidh) {
    http.Error(w, "500", 500)
    return
  }
  sid := bson.ObjectIdHex(sidh)
  var submission models.Submission
  Decode(&submission, r)
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  if err := db.C("submissions").Update(bson.M{"_id": sid}, bson.M{"$set": bson.M{
    "content": submission.Content,
    "lang": submission.Lang,
  }}); err != nil {
    log.Println("Update submission: ", err)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/status/" + sidh, 302)
}

// GET /problems/:id/status
func ProblemStatusHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Error(w, "500", 500)
    return
  }
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  it := db.C("submissions").Find(bson.M{"pid": pid}).Sort("-_id").Limit(50).Iter()
  var submissions []models.Submission
  if err := it.All(&submissions); err != nil {
    log.Println("Problem status index: ", err)
  }
  render("status/index.html", c, w, submissions)
}

// GET /problems/:id/submit
func ProblemSubmitHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Error(w, "500", 500)
    return
  }
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  n, _ := db.C("problems").Find(bson.M{"_id": pid}).Count()
  if n == 0 {
    http.Error(w, "500", 500)
    return
  }
  render("status/new.html", c, w, pid)
}

// POST /problems/:id/submit
func ProblemSubmitHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  pid, err := strconv.Atoi(pat.Param(c, "id"))
  if err != nil {
    http.Error(w, "500", 500)
    return
  }
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  n, _ := db.C("problems").Find(bson.M{"_id": pid}).Count()
  if n == 0 {
    http.Error(w, "500", 500)
    return
  }
  user := CurrentUser(c)
  var submission models.Submission
  Decode(&submission, r)
  submission.ID = bson.NewObjectId()
  submission.Pid = pid
  submission.Username = user.Username
  submission.Verdict = models.Verdict{models.QUEUED, 0, 0}
  if !inRange(submission.Lang, 0, models.LANGSIZE-1) {
    submission.Lang = models.LANGCPP
  }
  if err := db.C("submissions").Insert(submission); err != nil {
    log.Println("Problem submit: ", err)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/status/" + submission.ID.Hex(), 302)
}

