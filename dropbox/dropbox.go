package dropbox

import (
//	"io"
	"net/http"
	"google.golang.org/appengine"
	log "google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/jpg0/dropbox"
	"io"
	"golang.org/x/oauth2"
//	"encoding/json"
	"store"
)

const DROPBOX_OAUTH_KEY string = "b0fx8jxeynzmoqd"
const DROPBOX_OAUTH_SECRET string = "x4fstmjx1b1yti5"

func NewDropboxClient(c context.Context) (*dropbox.Dropbox, *oauth2.Config) {
	rv := dropbox.NewDropbox()

	config := &oauth2.Config{
		ClientID:     DROPBOX_OAUTH_KEY,
		ClientSecret: DROPBOX_OAUTH_SECRET,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}

	rv.SetOAuth2Config(config)
	rv.SetAppInfo(DROPBOX_OAUTH_KEY, DROPBOX_OAUTH_SECRET)
	rv.SetContext(c)

	//config.

	return rv, config
}

func ConfigureDropbox(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	db, config := NewDropboxClient(c)

	var callback_url string
	if(r.Host[:9] == "localhost") {
		callback_url = "https://d2f-transfer.appspot.com" + r.URL.Path + "/callback"
	} else {
		callback_url = r.Host + ":/" + r.URL.Path + "/callback"
	}

	config.RedirectURL = callback_url

	//o, _ := json.Marshal(db.Config)

	log.Infof(c, "config=%v", config)
	log.Infof(c, "db.config=%v", db.Config)


	redirectUrl := db.AuthURL();

	log.Infof(c, "config.RedirectURL=%v", config.RedirectURL)

	http.Redirect(w, r, redirectUrl, 302)
}

func StoreDropboxConfiguration(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	code := r.URL.Query().Get("code")

	_, config := NewDropboxClient(c)

	token, err := config.Exchange(c, code)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = store.Save("dropbox", "access_token", token, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func OpenStreamForFile(title string, c context.Context) (io.ReadCloser, int64, error) {
	db, _ := NewDropboxClient(c)


	token := new(oauth2.Token)
	err := store.Load("dropbox", "access_token", token, c)

	if err != nil {
		return nil, 0, err
	}

	db.SetAccessToken(token.AccessToken)
	return db.Download(title, "", 0)
}