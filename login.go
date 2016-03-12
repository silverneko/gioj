package main

import (
  "net/http"
  "gopkg.in/mgo.v2/bson"
  "log"
  "golang.org/x/crypto/bcrypt"
)

type AuthCredential struct {
  Username string   `schema:"username"`
  Password string   `schema:"password"`
  Confirm_password string   `schema:"confirm_password"`
}

func LoginHandler(w http.ResponseWriter, req * http.Request) {
  render("login_form.html", w, "")
}

func AuthHandler(w http.ResponseWriter, req * http.Request) {
  var a AuthCredential
  Decode(&a, req)
  username, password := a.Username, a.Password

  db := DBSession{DB.Copy()}
  defer db.Close()
  var result map[string]string
  err := db.C("users").Find(bson.M{"username": username}).One(&result)
  if err != nil {
    /* User not found */
    render("login_form.html", w, "", "Username not found!")
    return
  }
  hashed := result["hashed_password"]
  ok := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
  if ok != nil {
    /* password mismatch */
    render("login_form.html", w, "", "Wrong password!")
    return
  }

  session, _ := cookieJar.Get(req, AuthSession)
  session.Values["username"] = username
  session.Save(req, w)
  http.Redirect(w, req, "/", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, req * http.Request) {
  session, _ := cookieJar.Get(req, AuthSession)
  session.Options.MaxAge = -1;
  session.Save(req, w)
  http.Redirect(w, req, "/", 302)
}

func RegisterHandler(w http.ResponseWriter, req * http.Request) {
  render("register_form.html", w, "")
}

func RegisterHandlerP(w http.ResponseWriter, req * http.Request) {
  var a AuthCredential
  Decode(&a, req)
  username, password, confirm := a.Username, a.Password, a.Confirm_password
  if len(username) < 3 {
    render("register_form.html", w, "", "Username is too short!")
    return
  }
  if len(password) < 8 {
    render("register_form.html", w, "", "Password is too short!")
    return
  }
  if password != confirm {
    render("register_form.html", w, "", "Confirm password mismatch!")
    return
  }
  hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
  db := DBSession{DB.Copy()}
  defer db.Close()
  err := db.C("users").Insert(bson.M{"username": username, "hashed_password": hashed})
  if err != nil {
    render("register_form.html", w, "", "Cannot use this username!")
    return
  }
  log.Println("Create user:", username)
  cookie, _ := cookieJar.Get(req, AuthSession)
  cookie.Values["username"] = username
  cookie.Save(req, w)
  http.Redirect(w, req, "/", 302)
}

