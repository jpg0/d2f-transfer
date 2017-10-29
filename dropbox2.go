package d2f_transfer

import (
	"time"
	"fmt"
	"golang.org/x/oauth2"
	"io"
)
import "github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"golang.org/x/net/context"
	"github.com/juju/errors"
)

func newConfig(t *oauth2.Token, c context.Context) dropbox.Config {

	timeoutCtx, _ := context.WithTimeout(c, 1*time.Minute)

	var conf = &oauth2.Config{Endpoint: dropbox.OAuthEndpoint(".dropboxapi.com")}
	client := conf.Client(timeoutCtx, t)


	return dropbox.Config{
		Token: t.AccessToken,
		LogLevel: dropbox.LogInfo,
		Client: client,
	}
}

func OpenStreamForFile(title string, validations *Validations, recentChange bool, c context.Context) (io.ReadCloser, uint64, error) {

	token := new(oauth2.Token)
	err := Load("dropbox", "access_token", token, c)

	if err != nil {
		return nil, 0, err
	}

	config := newConfig(token, c)

	db := files.New(config)

	if validations != nil { //we need to validate something

		err = validate(config, title, validations)

		if err != nil && recentChange { //wait and retry
			time.Sleep(5 * time.Second)
			err = validate(config, title, validations)
		}

		if err != nil {
			return nil, 0, err
		}
	}

	meta, content, err := db.Download(files.NewDownloadArg(title))

	return content, meta.Size, err


}

func validate(config dropbox.Config, title string, validations *Validations) (error) {

	db := files.New(config)


	meta, err := db.GetMetadata(&files.GetMetadataArg{Path:title})

	if err != nil {
		return err
	}

	fileMeta, ok := meta.(*files.FileMetadata)

	if !ok {
		return errors.Errorf("%v is not a file", title)
	}

	requestedTime := time.Time(*validations.Mtime).UTC()
	expectedTime := time.Time(fileMeta.ClientModified).UTC()

	if validations.Mtime != nil && !requestedTime.Equal(expectedTime) {
		return errors.New(fmt.Sprintf("Mtime mismatch: requested %v, actual %v: difference: %v", requestedTime, expectedTime, expectedTime.Sub(requestedTime)))
	}

	if validations.Size != nil && fileMeta.Size != *validations.Size {
		return errors.New(fmt.Sprintf("Size mismatch: requested %v, actual %v", *validations.Size, fileMeta.Size))
	}

	return nil
}
