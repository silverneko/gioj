package main

import (
  "net/http"
  "github.com/gorilla/sessions"
  "github.com/gorilla/schema"
  "encoding/hex"
  "gopkg.in/mgo.v2"
  "log"
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

var cookieJar = sessions.NewCookieStore(
  d("82468eae8c7b5428ac39b3de4b519f12c4b5927fbf9816c88242d04d6a2091f4"),
  d("e087ce5f3bbae7108d7808da9e64aff75d2fa9473ba91c80bc4afd1c6c094c7b"),
)

var DB *mgo.Session

func init() {
  cookieJar.Options = &sessions.Options{
    MaxAge:   86400 * 14,
    HttpOnly: true,
  }

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

  err = EnsureDBIndices()
  if err != nil {
    log.Fatal("mgo.EnsureIndex: ", err)
  }
}

func d(s string) []byte {
  b, _ := hex.DecodeString(s)
  return b
}

