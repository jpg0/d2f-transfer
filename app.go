package main

import (
	"net/http"
	"encoding/json"
//	"io"
//	"os"

//	"google.golang.org/appengine"

	"github.com/gorilla/mux"
	//"github.com/stacktic/dropbox"
	"google.golang.org/appengine"
	"flickr"
	"dropbox"
)


type TransferRequest struct {
	Title   string
	Tags    []string
	IsPublic, IsFamily, IsFriend bool
}


func init() {
	r := mux.NewRouter()
	r.HandleFunc("/transfer", Transfer)
	r.HandleFunc("/configure/flickr", flickr.ConfigureFlickr)
	r.HandleFunc("/configure/flickr/callback", flickr.StoreFlickrConfiguration)
	r.HandleFunc("/configure/dropbox", dropbox.ConfigureDropbox)
	r.HandleFunc("/configure/dropbox/callback", dropbox.StoreDropboxConfiguration)
	http.Handle("/", r)
}

func Transfer(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	transferRequest := new(TransferRequest)

	json.NewDecoder(r.Body).Decode(transferRequest)

	readCloser, _, err := dropbox.OpenStreamForFile(transferRequest.Title, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = flickr.Upload(
		transferRequest.Title,
		transferRequest.Tags,
		transferRequest.IsPublic,
		transferRequest.IsFamily,
		transferRequest.IsFriend,
		readCloser,
		c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

//	if u := user.Current(c); u != nil {
//		g.Author = u.String()
//	}


	w.WriteHeader(http.StatusNoContent)
}

