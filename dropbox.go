package d2f_transfer

import (
	"net/http"
	"google.golang.org/appengine"
	//log "google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/jpg0/dropbox"
	"io"
	"golang.org/x/oauth2"
)

const DROPBOX_OAUTH_KEY string = "b0fx8jxeynzmoqd"
const DROPBOX_OAUTH_SECRET string = "x4fstmjx1b1yti5"
const DROPBOX_CALLBACK_HOST string = "https://d2f-transfer.appspot.com"

func NewDropboxClient(c context.Context) (*dropbox.Dropbox, *oauth2.Config) {
	rv := dropbox.NewDropbox()

	rv.SetContext(c)
	config := &oauth2.Config{}

	rv.SetOAuth2Config(config)

	config.ClientID = DROPBOX_OAUTH_KEY
	config.ClientSecret = DROPBOX_OAUTH_SECRET
	config.Endpoint = oauth2.Endpoint{
		AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
		TokenURL: "https://api.dropbox.com/1/oauth2/token",
	}
	config.RedirectURL = DROPBOX_CALLBACK_HOST  + "/configure/dropbox/callback"

	return rv, config
}

func ConfigureDropbox(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	_, config := NewDropboxClient(c)

	redirectUrl := config.AuthCodeURL("");

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

	err = Save("dropbox", "access_token", token, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func OpenStreamForFile(title string, c context.Context) (io.ReadCloser, int64, error) {
	db, _ := NewDropboxClient(c)


	token := new(oauth2.Token)
	err := Load("dropbox", "access_token", token, c)

	if err != nil {
		return nil, 0, err
	}

	db.SetAccessToken(token.AccessToken)
	return db.Download(title, "", 0)
}