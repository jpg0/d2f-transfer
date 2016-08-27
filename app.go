package d2f_transfer

import (
	"net/http"
	"encoding/json"
//	"io"
//	"os"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	//"gopkg.in/masci/flickr.v2"
)


type TransferRequest struct {
	Title   string
	Tags    []string
	IsPublic, IsFamily, IsFriend bool
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/transfer", Transfer)
	r.HandleFunc("/configure/flickr", ConfigureFlickr)
	r.HandleFunc("/configure/flickr/callback", StoreFlickrConfiguration)
	r.HandleFunc("/configure/dropbox", ConfigureDropbox)
	r.HandleFunc("/configure/dropbox/callback", StoreDropboxConfiguration)
	r.HandleFunc("/test", TestStorage)
	http.Handle("/", r)
}

func TestStorage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := Save("test_namespace", "test_key", "test_value", c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Transfer(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	transferRequest := new(TransferRequest)

	json.NewDecoder(r.Body).Decode(transferRequest)

	readCloser, _, err := OpenStreamForFile(transferRequest.Title, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = Upload(
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

