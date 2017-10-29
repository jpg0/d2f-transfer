package d2f_transfer

import (
	"net/http"
	"google.golang.org/appengine"
	"golang.org/x/oauth2"
	"fmt"
)

const DROPBOX_OAUTH_KEY string = "b0fx8jxeynzmoqd"
const DROPBOX_OAUTH_SECRET string = "x4fstmjx1b1yti5"
const DROPBOX_CALLBACK_HOST string = "https://d2f-transfer.appspot.com"

func NewOAuthConfig() (*oauth2.Config){
	config := &oauth2.Config{}
	config.ClientID = DROPBOX_OAUTH_KEY
	config.ClientSecret = DROPBOX_OAUTH_SECRET
	config.Endpoint = oauth2.Endpoint{
		AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
		TokenURL: "https://api.dropbox.com/1/oauth2/token",
	}
	config.RedirectURL = DROPBOX_CALLBACK_HOST  + "/configure/dropbox/callback"

	return config
}

func ConfigureDropboxOAuth(w http.ResponseWriter, r *http.Request) {
	config := NewOAuthConfig()
	redirectUrl := config.AuthCodeURL("")
	http.Redirect(w, r, redirectUrl, 302)
}

func StoreDropboxConfiguration(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	code := r.URL.Query().Get("code")

	config := NewOAuthConfig()

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

	fmt.Fprint(w, "Dropbox Auth Configured")
	w.WriteHeader(http.StatusOK)
}