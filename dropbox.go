package d2f_transfer

import (
	"net/http"
	"google.golang.org/appengine"
	//log "google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/jpg0/dropbox"
	"io"
	"golang.org/x/oauth2"
	"fmt"
	"time"
	"errors"
)

const DROPBOX_OAUTH_KEY string = "b0fx8jxeynzmoqd"
const DROPBOX_OAUTH_SECRET string = "x4fstmjx1b1yti5"
const DROPBOX_CALLBACK_HOST string = "https://d2f-transfer.appspot.com"

type Validations struct {
	Mtime *dropbox.DBTime
	Size *int64
}

func NewDropboxClient(c context.Context) (*dropbox.Dropbox, *oauth2.Config) {
	rv := dropbox.NewDropbox()

	timeoutCtx, _ := context.WithTimeout(c, 1*time.Minute)

	rv.SetContext(timeoutCtx)
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

	fmt.Fprint(w, "Dropbox Auth Configured")
	w.WriteHeader(http.StatusOK)
}

func OpenStreamForFile(title string, validations *Validations, recentChange bool, c context.Context) (io.ReadCloser, int64, error) {
	db, _ := NewDropboxClient(c)


	token := new(oauth2.Token)
	err := Load("dropbox", "access_token", token, c)

	if err != nil {
		return nil, 0, err
	}

	db.SetAccessToken(token.AccessToken)

	if validations != nil { //we need to validate something

		err = Validate(db, title, validations)

		if(err != nil && recentChange){ //wait and retry
			time.Sleep(5 * time.Second)
			err = Validate(db, title, validations)
		}

		if (err != nil) {
			return nil, 0, err
		}
	}

	return db.Download(title, "", 0)
}

func Validate(db *dropbox.Dropbox, title string, validations *Validations) (error) {
	entry, err := db.Metadata(title, false, false, "", "", 0)

	if err != nil {
		return err
	}

	requestedTime := time.Time(*validations.Mtime).UTC()
	expectedTime := time.Time(entry.ClientMtime).UTC()

	if validations.Mtime != nil && !requestedTime.Equal(expectedTime) {
		return errors.New(fmt.Sprintf("Mtime mismatch: requested %v, actual %v: difference: %v", requestedTime, expectedTime, expectedTime.Sub(requestedTime)))
	}

	if validations.Size != nil && entry.Bytes != *validations.Size {
		return errors.New(fmt.Sprintf("Size mismatch: requested %v, actual %v", *validations.Size, entry.Bytes))
	}

	return nil
}