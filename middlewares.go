package main

import (
  "net/http"
  "goji.io"
  "gopkg.in/mgo.v2/bson"
  "golang.org/x/net/context"
  "log"
  "bytes"
  "github.com/silverneko/gioj/models"
)

type Key int

const (
  userKey Key = iota
)

func AuthMiddleware(h goji.Handler) goji.Handler {
  return goji.HandlerFunc(
    func (c context.Context, w http.ResponseWriter, r *http.Request) {
      var user *models.User = nil
      var cookie []interface{}
      if err := cookieJar.GetCookie(r, AuthSession, &cookie); err != nil {
        /* Invalid cookie */
        log.Println(err)
        cookieJar.DestroyCookie(w, AuthSession)
      } else {
	username, ok1 := cookie[0].(string)
	hashed, ok2 := cookie[1].([]byte)
	if (!ok1) || (!ok2) {
	  log.Println("Invalid cookie")
	  cookieJar.DestroyCookie(w, AuthSession)
	}
        db := models.DBSession{DB.Copy()}
        defer db.Close()
        var result models.User
        if err := db.C("users").Find(bson.M{"username": username}).One(&result); err != nil {
          /* username don't exist in db */
          log.Println(err)
          cookieJar.DestroyCookie(w, AuthSession)
        } else {
	  if !bytes.Equal(result.Hashed_password, hashed) {
	    log.Println("Invalid cookie password hashsum:", username)
	    cookieJar.DestroyCookie(w, AuthSession)
	  } else {
	    user = &result
	  }
        }
      }
      ctx := context.WithValue(c, userKey, user)
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

func CurrentUser(c context.Context) *models.User {
  user := c.Value(userKey).(*models.User)
  return user
}

func isLogin(c context.Context) bool {
  return CurrentUser(c) != nil
}

