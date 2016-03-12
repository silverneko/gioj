package main

import (
  "gopkg.in/mgo.v2"
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
