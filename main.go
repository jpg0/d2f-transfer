package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"flickr"
	"dropbox"
)

func Main(args []string) {
	r := mux.NewRouter()
	r.HandleFunc("/transfer", Transfer)
	r.HandleFunc("/configure/flickr", flickr.ConfigureFlickr)
	r.HandleFunc("/configure/flickr/callback", flickr.StoreFlickrConfiguration)
	r.HandleFunc("/configure/dropbox", dropbox.ConfigureDropbox)
	r.HandleFunc("/configure/dropbox/callback", dropbox.StoreDropboxConfiguration)
	http.Handle("/", r)
}