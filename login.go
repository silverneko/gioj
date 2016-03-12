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
    log.Println(err)
    http.Redirect(w, req, "/login", http.StatusFound)
    return
  }
  hashed := result["hashed_password"]
  ok := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
  if ok != nil {
    /* password mismatch */
    log.Println(ok)
    http.Redirect(w, req, "/login", http.StatusFound)
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
  if password != confirm {
    http.Redirect(w, req, "/register", 302)
    return
  }
  hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
  dbConnection := DB.Copy()
  defer dbConnection.Close()
  err := dbConnection.DB("").C("users").Insert(bson.M{"username": username, "hashed_password": hashed})
  if err != nil {
    http.Redirect(w, req, "/register", 302)
    return
  }
  cookie, _ := cookieJar.Get(req, AuthSession)
  cookie.Values["username"] = username
  cookie.Save(req, w)
  http.Redirect(w, req, "/", 302)
}

