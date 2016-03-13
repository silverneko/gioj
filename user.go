package main

import (
  "net/http"
  "goji.io/pat"
  "golang.org/x/net/context"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/crypto/bcrypt"
  "log"
)

func UserHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  username := pat.Param(c, "user")
  db := DBSession{DB.Copy()}
  defer db.Close()
  var result User
  err := db.C("users").Find(bson.M{"username": username}).One(&result)
  if err != nil {
    /* username don't exist in db */
    http.Error(w, "500", 500)
    return
  }
  render("user/show.html", w, r, result);
}

func UserEditHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  user := CurrentUser(w, r)
  if user.Username != pat.Param(c, "user") {
    http.Error(w, "500", 500)
    return
  }
  render("user/edit_form.html", w, r, "")
}

type UserEditForm struct {
  Name string
  New_password string
  Confirm_password string
  Old_password string
}

func UserEditHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  user := CurrentUser(w, r)
  if user.Username != pat.Param(c, "user") {
    http.Redirect(w, r, "/", 302)
    return
  }
  var form UserEditForm
  Decode(&form, r)
  name, newpwd, confirmpwd, oldpwd := form.Name, form.New_password, form.Confirm_password, form.Old_password
  ok := bcrypt.CompareHashAndPassword(user.Hashed_password, []byte(oldpwd))
  if ok != nil {
    /* password mismatch */
    render("user/edit_form.html", w, r, "", "Wrong password!")
    return
  }
  if len(name) >= 3 {
    user.Name = name
  }
  if len(newpwd) >= 8 {
    if newpwd != confirmpwd {
      render("user/edit_form.html", w, r, "", "Confirm new password mismatch!")
      return
    }
    hashed, _ := bcrypt.GenerateFromPassword([]byte(newpwd), 10)
    user.Hashed_password = hashed
  }
  db := DBSession{DB.Copy()}
  defer db.Close()
  err := db.C("users").Update(bson.M{"_id": user.ID}, user)
  if err != nil {
    log.Println("User update: ", err, user.ID)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/", 302)
}
