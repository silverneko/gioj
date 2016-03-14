package main

import (
  "net/http"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "log"
)

type DBSession struct {
  *mgo.Session
}

func (s DBSession) C(name string) *mgo.Collection {
  return s.DB("").C(name)
}

func EnsureDBIndices() error {
  err := DB.DB("").C("users").EnsureIndex(mgo.Index{
    Key: []string{"username"},
    Unique: true,
  })
  if err != nil {
    return err
  }
  return nil
}

type DiscussPost struct {
  ID bson.ObjectId   `bson:"_id"`
  Content string
  Username string
}

type User struct {
  ID bson.ObjectId   `bson:"_id"`
  Name string
  Username string
  Hashed_password []byte
}

func CurrentUser(w http.ResponseWriter, r *http.Request) *User {
  session, err := cookieJar.Get(r, AuthSession)
  if err != nil {
    /* Invalid cookie */
    log.Println(err)
    session.Options.MaxAge = -1
    session.Save(r, w)
    return nil
  }
  username, ok := session.Values["username"].(string)
  if !ok || username == ""{
    return nil
  }
  db := DBSession{DB.Copy()}
  defer db.Close()
  var result User
  err = db.C("users").Find(bson.M{"username": username}).One(&result)
  if err != nil {
    /* username don't exist in db */
    log.Println(err)
    session.Options.MaxAge = -1
    session.Save(r, w)
    return nil
  }
  return &result
}
