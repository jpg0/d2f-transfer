package d2f_transfer

import (
	"net/http"
	"github.com/gorilla/mux"
)

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/transfer", use(Transfer, CanTransfer))
	r.HandleFunc("/configure/flickr", use(ConfigureFlickr, CanConfigure))
	r.HandleFunc("/configure/flickr/callback", use(StoreFlickrConfiguration, CanConfigure))
	r.HandleFunc("/configure/dropbox", use(ConfigureDropboxOAuth, CanConfigure))
	r.HandleFunc("/configure/dropbox/callback", use(StoreDropboxConfiguration, CanConfigure))
	http.Handle("/", r)
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}