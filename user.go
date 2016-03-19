package main

import (
  "net/http"
  "goji.io/pat"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/net/context"
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
  if err := db.C("users").Find(bson.M{"username": username}).One(&result); err != nil {
    /* username don't exist in db */
    http.Error(w, "500", 500)
    return
  }
  render("user/show.html", c, w, result);
}

func UserEditHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  user := c.Value("currentUser").(*User)
  if user.Username != pat.Param(c, "user") {
    http.Error(w, "500", 500)
    return
  }
  render("user/edit_form.html", c, w, "")
}

type UserEditForm struct {
  Name string
  New_password string
  Confirm_password string
  Old_password string
}

func UserEditHandlerP(c context.Context, w http.ResponseWriter, r *http.Request) {
  user := c.Value("currentUser").(*User)
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
      render("user/edit_form.html", c, w, "", "Wrong password!")
      return
    }
  } else {
    http.Error(w, "500", 500)
    return
  }
  if name != "" {
    if inRange(len(name), 3, 20) {
      user.Name = name
    } else {
      render("user/edit_form.html", c, w, "", "Nickname length: 3 ~ 20!")
      return
    }
  }
  if newpwd != "" {
    if inRange(len(newpwd), 8, 40) {
      if newpwd == confirmpwd {
        hashed, _ := bcrypt.GenerateFromPassword([]byte(newpwd), 10)
        user.Hashed_password = hashed
      } else {
        render("user/edit_form.html", c, w, "", "Confirm new password mismatch!")
        return
      }
    } else {
      render("user/edit_form.html", c, w, "", "Password length: 8 ~ 40!")
      return
    }
  }
  db := DBSession{DB.Copy()}
  defer db.Close()
  if err := db.C("users").Update(bson.M{"_id": user.ID}, user); err != nil {
    log.Println("User update: ", err, user.ID)
    http.Error(w, "500", 500)
    return
  }
  log.Println("User update: ", user)
  http.Redirect(w, r, "/", 302)
}

func LoginHandler(c context.Context, w http.ResponseWriter, r * http.Request) {
  if isLogin(c) {
    http.Redirect(w, r, "/", 302)
    return
  }
  render("user/login_form.html", c, w, "")
}

func LogoutHandler(c context.Context, w http.ResponseWriter, r * http.Request) {
  cookieJar.DestroyCookie(w, AuthSession)
  http.Redirect(w, r, "/", 302)
}

type AuthCredential struct {
  Name string	    `schema:"name"`
  Username string   `schema:"username"`
  Password string   `schema:"password"`
  Confirm_password string   `schema:"confirm_password"`
}

func AuthHandler(c context.Context, w http.ResponseWriter, r * http.Request) {
  if isLogin(c) {
    http.Redirect(w, r, "/", 302)
    return
  }
  var a AuthCredential
  Decode(&a, r)
  username, password := a.Username, a.Password
  if !inRange(len(username), 3, 20) {
    render("user/login_form.html", c, w, "", "Username not found!")
    return
  }
  if !inRange(len(password), 8, 40) {
    render("user/login_form.html", c, w, "", "Username not found!")
    return
  }

  db := DBSession{DB.Copy()}
  defer db.Close()
  var result User
  if err := db.C("users").Find(bson.M{"username": username}).One(&result); err != nil {
    /* User not found */
    render("user/login_form.html", c, w, "", "Username not found!")
    return
  }
  ok := bcrypt.CompareHashAndPassword(result.Hashed_password, []byte(password))
  if ok != nil {
    /* password mismatch */
    render("user/login_form.html", c, w, "", "Wrong password!")
    return
  }
  log.Println("User login: ", username)
  cookieJar.PutCookie(w, AuthSession, username)
  http.Redirect(w, r, "/", http.StatusFound)
}

func RegisterHandler(c context.Context, w http.ResponseWriter, r * http.Request) {
  if isLogin(c) {
    http.Redirect(w, r, "/", 302)
    return
  }
  render("user/register_form.html", c, w, "")
}

func RegisterHandlerP(c context.Context, w http.ResponseWriter, r * http.Request) {
  if isLogin(c) {
    http.Redirect(w, r, "/", 302)
    return
  }
  var a AuthCredential
  Decode(&a, r)
  name, username, password, confirm := a.Name, a.Username, a.Password, a.Confirm_password
  if !inRange(len(name), 3, 20) {
    render("user/register_form.html", c, w, "", "Nickname length: 3 ~ 20!")
    return
  }
  if !inRange(len(username), 3, 20) {
    render("user/register_form.html", c, w, "", "Username length: 3 ~ 20!")
    return
  }
  if !inRange(len(password), 8, 40) {
    render("user/register_form.html", c, w, "", "Password length: 8 ~ 40!")
    return
  }
  if password != confirm {
    render("user/register_form.html", c, w, "", "Confirm password mismatch!")
    return
  }
  hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
  db := DBSession{DB.Copy()}
  defer db.Close()
  if err := db.C("users").Insert(bson.M{"name": name, "username": username, "hashed_password": hashed}); err != nil {
    render("user/register_form.html", c, w, "", "Cannot use this username!")
    return
  }
  log.Println("Create user: ", username)
  cookieJar.PutCookie(w, AuthSession, username)
  http.Redirect(w, r, "/", 302)
}

