package d2f_transfer

import (
	"io"
	"golang.org/x/oauth2"
	"time"
	"fmt"
	"github.com/jpg0/dropbox"
	"golang.org/x/net/context"
	"errors"
)

type Validations struct {
	Mtime *dropbox.DBTime
	Size *uint64
}

func newDropboxClient(c context.Context, config *oauth2.Config) (*dropbox.Dropbox) {
	rv := dropbox.NewDropbox()

	timeoutCtx, _ := context.WithTimeout(c, 1*time.Minute)

	rv.SetContext(timeoutCtx)
	rv.SetOAuth2Config(config)

	return rv
}

func OpenStreamForFileV1(title string, validations *Validations, recentChange bool, c context.Context) (io.ReadCloser, uint64, error) {
	db := newDropboxClient(c, NewOAuthConfig())


	token := new(oauth2.Token)
	err := Load("dropbox", "access_token", token, c)

	if err != nil {
		return nil, 0, err
	}

	db.SetAccessToken(token.AccessToken)

	if validations != nil { //we need to validate something

		err = validate1(db, title, validations)

		if(err != nil && recentChange){ //wait and retry
			time.Sleep(5 * time.Second)
			err = validate1(db, title, validations)
		}

		if (err != nil) {
			return nil, 0, err
		}
	}

	rc, size, err := db.Download(title, "", 0)

	return rc, uint64(size), err
}

func validate1(db *dropbox.Dropbox, title string, validations *Validations) (error) {
	entry, err := db.Metadata(title, false, false, "", "", 0)

	if err != nil {
		return err
	}

	requestedTime := time.Time(*validations.Mtime).UTC()
	expectedTime := time.Time(entry.ClientMtime).UTC()

	if validations.Mtime != nil && !requestedTime.Equal(expectedTime) {
		return errors.New(fmt.Sprintf("Mtime mismatch: requested %v, actual %v: difference: %v", requestedTime, expectedTime, expectedTime.Sub(requestedTime)))
	}

	if validations.Size != nil && uint64(entry.Bytes) != *validations.Size {
		return errors.New(fmt.Sprintf("Size mismatch: requested %v, actual %v", *validations.Size, entry.Bytes))
	}

	return nil
}
