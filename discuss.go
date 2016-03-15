package main

import (
  "net/http"
  "golang.org/x/net/context"
  "gopkg.in/mgo.v2/bson"
  "log"
)

// GET /discuss
func DiscussHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  if !isLogin(c) {
    http.Redirect(w, r, "/login", 302)
    return
  }
  db := DBSession{DB.Copy()}
  defer db.Close()
  it := db.C("discuss").Find(nil).Sort("-_id").Limit(50).Iter()
  var discussPosts []DiscussPost
  if err := it.All(&discussPosts); err != nil {
    log.Println("Discuss index: ", err)
  }
  render("discuss/index.html", c, w, discussPosts)
}

// POST /discuss/new
func DiscussNewHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
  if !isLogin(c) {
    http.Redirect(w, r, "/login", 302)
    return
  }
  var post DiscussPost
  Decode(&post, r)
  if post.Content == "" {
    // Do nothing
    http.Redirect(w, r, "/discuss", 302)
    return
  }
  if len(post.Content) > 512 {
    post.Content = post.Content[:512]
  }
  user := c.Value("currentUser").(*User)
  db := DBSession{DB.Copy()}
  defer db.Close()
  err := db.C("discuss").Insert(bson.M{"content": post.Content, "username": user.Username})
  if err != nil {
    log.Println("New discuss post: ", err)
    http.Error(w, "500", 500)
    return
  }
  http.Redirect(w, r, "/discuss", 302)
}

