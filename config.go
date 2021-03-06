package main

import (
  "net/http"
  "github.com/gorilla/securecookie"
  "github.com/gorilla/schema"
  "encoding/hex"
  "gopkg.in/mgo.v2"
  "log"
  "os"
  "github.com/silverneko/gioj/models"
)

var decoder = schema.NewDecoder()

func Decode(dst interface{}, r *http.Request) error {
  if err := r.ParseForm(); err != nil {
    return err
  }
  if err := decoder.Decode(dst, r.PostForm); err != nil {
    return err
  }
  return nil
}

const AuthSession = "gioj-auth"
var cookieJar CookieJar

var DB *mgo.Session
var GIOJROOT string

func init() {
  cookieJar = CookieJar{
    MaxAge:   86400 * 7,
    HttpOnly: true,
    Secure:   false,
    SecureCookie: securecookie.New(
      d("82468eae8c7b5428ac39b3de4b519f12c4b5927fbf9816c88242d04d6a2091f4"),
      d("e087ce5f3bbae7108d7808da9e64aff75d2fa9473ba91c80bc4afd1c6c094c7b"),
    ),
  }

  /*
  mgo.SetDebug(true)
  mgo.SetLogger(log.New(os.Stderr, "[Database]", log.LstdFlags))
  */
  var err error
  DB, err = mgo.DialWithInfo(&mgo.DialInfo{
    Addrs: []string{"localhost"},
    Database: "gioj_test",
    Username: "gioj",
    Password: "gioj",
  })
  if err != nil {
    log.Fatal("mgo.Dial: ", err)
  }

  if err := models.EnsureDBIndices(DB); err != nil {
    log.Fatal("mgo.EnsureIndex: ", err)
  }

  GIOJROOT = os.Getenv("GOPATH") + "/src/github.com/silverneko/gioj"
}

func d(s string) []byte {
  b, _ := hex.DecodeString(s)
  return b
}

