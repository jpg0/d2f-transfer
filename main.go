package d2f_transfer

import (
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	r := mux.NewRouter()
//	r.HandleFunc("/transfer", Transfer)
	r.HandleFunc("/configure/flickr", ConfigureFlickr)
	r.HandleFunc("/configure/flickr/callback", StoreFlickrConfiguration)
	r.HandleFunc("/configure/dropbox", ConfigureDropbox)
	r.HandleFunc("/configure/dropbox/callback", StoreDropboxConfiguration)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}