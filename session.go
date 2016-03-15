package main

import (
  "net/http"
  "time"
  "github.com/gorilla/securecookie"
)

type CookieJar struct {
  MaxAge int
  Secure bool
  HttpOnly bool
  Path string
  Domain string
  *securecookie.SecureCookie
}

func (cjar *CookieJar) GetCookie(r *http.Request, name string, result interface{}) error {
  cookie, err := r.Cookie(name)
  if err != nil {
    // Not found
    return err
  }
  err = cjar.Decode(name, cookie.Value, result)
  if err != nil {
    return err
  }
  return nil
}

func (cjar *CookieJar) PutCookie(w http.ResponseWriter, name string, result interface{}) error {
  encoded, err := cjar.Encode(name, result)
  if err != nil {
    return err
  }
  cookie := &http.Cookie{
    Name: name,
    Value: encoded,
    MaxAge: cjar.MaxAge,
    HttpOnly: cjar.HttpOnly,
    Secure: cjar.Secure,
    Path: cjar.Path,
    Domain: cjar.Domain,
  }
  if cookie.MaxAge > 0 {
    d := time.Duration(cookie.MaxAge) * time.Second
    cookie.Expires = time.Now().Add(d)
  } else if cookie.MaxAge < 0 {
    cookie.Expires = time.Unix(1, 0)
  }
  http.SetCookie(w, cookie)
  return nil
}

func (cjar *CookieJar) DestroyCookie(w http.ResponseWriter, name string) {
  cookie := &http.Cookie{
    Name: name,
    Value: "",
    MaxAge: -1,
    HttpOnly: cjar.HttpOnly,
    Secure: cjar.Secure,
    Path: cjar.Path,
    Domain: cjar.Domain,
    Expires: time.Unix(1, 0),
  }
  http.SetCookie(w, cookie)
}

