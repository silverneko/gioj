package main

import (
  "net/http"
  "goji.io"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/net/context"
  "log"
)

func AuthMiddleware(h goji.Handler) goji.Handler {
  return goji.HandlerFunc(
    func (c context.Context, w http.ResponseWriter, r *http.Request) {
      var user *User = nil
      var username string
      if err := cookieJar.GetCookie(r, AuthSession, &username); err != nil {
        /* Invalid cookie */
        log.Println(err)
        cookieJar.DestroyCookie(w, AuthSession)
      } else if username == "" {
        cookieJar.DestroyCookie(w, AuthSession)
      } else {
        db := DBSession{DB.Copy()}
        defer db.Close()
        var result User
        if err := db.C("users").Find(bson.M{"username": username}).One(&result); err != nil {
          /* username don't exist in db */
          log.Println(err)
          cookieJar.DestroyCookie(w, AuthSession)
        } else {
          user = &result
        }
      }
      ctx := context.WithValue(c, "currentUser", user)
      h.ServeHTTPC(ctx, w, r)
    },
  )
}

func RequireAuth(h goji.HandlerFunc) goji.Handler {
  return goji.HandlerFunc(
    func (c context.Context, w http.ResponseWriter, r *http.Request) {
      if !isLogin(c) {
	http.Redirect(w, r, "/login", 302)
	return
      } else {
	h.ServeHTTPC(c, w, r)
      }
    },
  )
}

func isLogin(c context.Context) bool {
  user := c.Value("currentUser").(*User)
  return user != nil
}

