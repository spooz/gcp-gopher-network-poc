package main

import (
        "net/http"
        "fmt"
        "html/template"
        "time"


        "google.golang.org/appengine" // Required external App Engine library
        "google.golang.org/appengine/datastore"
        "google.golang.org/appengine/log"
)

var (
        indexTemplate = template.Must(template.ParseFiles("index.html"))
)
type templateParams struct {
        Notice string
        Name   string
        Message string
        Posts []Post
}

type Post struct {
        Author  string
        Message string
        Posted  time.Time
}

func main() {
        http.HandleFunc("/", indexHandler)
        appengine.Main() // Starts the server to receive requests
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
      	if r.URL.Path != "/" {
      		http.Redirect(w, r, "/", http.StatusFound)
      		return
      	}

      	ctx := appengine.NewContext(r)

      	params := templateParams{}

      	q := datastore.NewQuery("Post").Order("-Posted").Limit(20)

      	if _, err := q.GetAll(ctx, &params.Posts); err != nil {
      		log.Errorf(ctx, "Getting posts: %v", err)
      		w.WriteHeader(http.StatusInternalServerError)
      		params.Notice = "Couldn't get latest posts. Refresh?"
      		indexTemplate.Execute(w, params)
      		return
      	}

      	if r.Method == "GET" {
      		indexTemplate.Execute(w, params)
      		return
      	}

      	post := Post{
      		Author:  r.FormValue("name"),
      		Message: r.FormValue("message"),
      		Posted:  time.Now(),
      	}

      	if post.Author == "" {
      		post.Author = "Anonymous Gopher"
      	}
      	params.Name = post.Author

      	if post.Message == "" {
      		w.WriteHeader(http.StatusBadRequest)
      		params.Notice = "No message provided"
      		indexTemplate.Execute(w, params)
      		return
      	}

      	key := datastore.NewIncompleteKey(ctx, "Post", nil)

      	if _, err := datastore.Put(ctx, key, &post); err != nil {
      		log.Errorf(ctx, "datastore.Put: %v", err)

      		w.WriteHeader(http.StatusInternalServerError)
      		params.Notice = "Couldn't add new post. Try again?"
      		params.Message = post.Message // Preserve their message so they can try again.
      		indexTemplate.Execute(w, params)
      		return
      	}
      	params.Posts = append([]Post{post}, params.Posts...)

      	params.Notice = fmt.Sprintf("Thank you for your submission, %s!", post.Author)
      	indexTemplate.Execute(w, params)
}
