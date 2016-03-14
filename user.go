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
  if !inRange(len(username), 3, 20) {
    http.Error(w, "500", 500)
    return
  }
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
  if inRange(len(oldpwd), 8, 40) {
    ok := bcrypt.CompareHashAndPassword(user.Hashed_password, []byte(oldpwd))
    if ok != nil {
      /* password mismatch */
      render("user/edit_form.html", w, r, "", "Wrong password!")
      return
    }
  } else {
    http.Error(w, "500", 500)
    return
  }
  if !inRange(len(name), 3, 20) {
    render("user/edit_form.html", w, r, "", "Nickname length: 3 ~ 20!")
    return
  }
  if !inRange(len(newpwd), 8, 40) {
    render("user/edit_form.html", w, r, "", "Password length: 8 ~ 40!")
    return
  }
  if newpwd != confirmpwd {
    render("user/edit_form.html", w, r, "", "Confirm new password mismatch!")
    return
  }
  hashed, _ := bcrypt.GenerateFromPassword([]byte(newpwd), 10)
  user.Hashed_password = hashed
  user.Name = name
  db := DBSession{DB.Copy()}
  defer db.Close()
  err := db.C("users").Update(bson.M{"_id": user.ID}, user)
  if err != nil {
    log.Println("User update: ", err, user.ID)
    http.Error(w, "500", 500)
    return
  }
  log.Println("User update: ", user)
  http.Redirect(w, r, "/", 302)
}

func LoginHandler(w http.ResponseWriter, req * http.Request) {
  render("user/login_form.html", w, req, "")
}

func LogoutHandler(w http.ResponseWriter, req * http.Request) {
  session, _ := cookieJar.Get(req, AuthSession)
  session.Options.MaxAge = -1;
  session.Save(req, w)
  http.Redirect(w, req, "/", 302)
}

type AuthCredential struct {
  Name string	    `schema:"name"`
  Username string   `schema:"username"`
  Password string   `schema:"password"`
  Confirm_password string   `schema:"confirm_password"`
}

func AuthHandler(w http.ResponseWriter, req * http.Request) {
  var a AuthCredential
  Decode(&a, req)
  username, password := a.Username, a.Password
  if !inRange(len(username), 3, 20) {
    render("user/login_form.html", w, req, "", "Username not found!")
    return
  }
  if !inRange(len(password), 8, 40) {
    render("user/login_form.html", w, req, "", "Username not found!")
    return
  }

  db := DBSession{DB.Copy()}
  defer db.Close()
  var result User
  err := db.C("users").Find(bson.M{"username": username}).One(&result)
  if err != nil {
    /* User not found */
    render("user/login_form.html", w, req, "", "Username not found!")
    return
  }
  ok := bcrypt.CompareHashAndPassword(result.Hashed_password, []byte(password))
  if ok != nil {
    /* password mismatch */
    render("user/login_form.html", w, req, "", "Wrong password!")
    return
  }
  log.Println("User login: ", username)
  session, _ := cookieJar.Get(req, AuthSession)
  session.Values["username"] = username
  session.Save(req, w)
  http.Redirect(w, req, "/", http.StatusFound)
}

func RegisterHandler(w http.ResponseWriter, req * http.Request) {
  render("user/register_form.html", w, req, "")
}

func RegisterHandlerP(w http.ResponseWriter, req * http.Request) {
  var a AuthCredential
  Decode(&a, req)
  name, username, password, confirm := a.Name, a.Username, a.Password, a.Confirm_password
  if !inRange(len(name), 3, 20) {
    render("user/register_form.html", w, req, "", "Nickname length: 3 ~ 20!")
    return
  }
  if !inRange(len(username), 3, 20) {
    render("user/register_form.html", w, req, "", "Username length: 3 ~ 20!")
    return
  }
  if !inRange(len(password), 8, 40) {
    render("user/register_form.html", w, req, "", "Password length: 8 ~ 40!")
    return
  }
  if password != confirm {
    render("user/register_form.html", w, req, "", "Confirm password mismatch!")
    return
  }
  hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
  db := DBSession{DB.Copy()}
  defer db.Close()
  err := db.C("users").Insert(bson.M{"name": name, "username": username, "hashed_password": hashed})
  if err != nil {
    render("user/register_form.html", w, req, "", "Cannot use this username!")
    return
  }
  log.Println("Create user: ", username)
  cookie, _ := cookieJar.Get(req, AuthSession)
  cookie.Values["username"] = username
  cookie.Save(req, w)
  http.Redirect(w, req, "/", 302)
}

