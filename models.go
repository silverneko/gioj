package main

import (
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
    log.Println(err)
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

