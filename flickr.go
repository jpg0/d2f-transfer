package d2f_transfer

import (
	"io"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	log "google.golang.org/appengine/log"

	"gopkg.in/masci/flickr.v2"
	"golang.org/x/net/context"
	"io/ioutil"
	"errors"
	//"github.com/jpg0/d2f-transfer/store"
)

const FLICKR_OAUTH_KEY string = "5845f87d43f2fa6ca0b328272dbf9395"
const FLICKR_OAUTH_SECRET string = "f45c95fe9515ec77"

func NewFlickrClient(c context.Context)(*flickr.FlickrClient) {
	client := flickr.NewFlickrClient(FLICKR_OAUTH_KEY, FLICKR_OAUTH_SECRET)

	//override client to work with GAE
	client.HTTPClient = urlfetch.Client(c)

	return client
}

// Retrieve a request token: this is the first step to get a fully functional
// access token from Flickr
func GetRequestTokenWithCallback(client *flickr.FlickrClient, callback string) (*flickr.RequestToken, error) {
	client.EndpointUrl = flickr.REQUEST_TOKEN_URL
	client.SetOAuthDefaults()
	client.Args.Set("oauth_consumer_key", client.ApiKey)
	client.Args.Set("oauth_callback", callback)

	// we don't have token secret at this stage, pass an empty string
	client.Sign("")

	res, err := client.HTTPClient.Get(client.GetUrl())
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return flickr.ParseRequestToken(string(body))
}

func ConfigureFlickr(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	client := NewFlickrClient(c)

	var callback_url string
	if(r.Host[:9] == "localhost") {
		callback_url = "https://d2f-transfer.appspot.com" + r.URL.Path + "/callback"
	} else {
		callback_url = r.Host + ":/" + r.URL.Path + "/callback"
	}
	// first, get a request token
	requestTok, err := GetRequestTokenWithCallback(client, callback_url)


	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// build the authorizatin URL
	url, _ := flickr.GetAuthorizeUrl(client, requestTok)

	//switch delete for write
	newUrl := url[:len(url) - 6] + "write"

	err = Save("flickr", "request_token", requestTok, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, newUrl, 302)

	// ask user to hit the authorization url with
	// their browser, authorize this application and coming
	// back with the confirmation token

	// finally, get the access token, setup the client and start making requests
//	accessTok, err := flickr.GetAccessToken(client, requestTok, "oauth_confirmation_code")
//	client.OAuthToken = accessTok.OAuthToken
//	client.OAuthTokenSecret = accessTok.OAuthTokenSecret
}



func StoreFlickrConfiguration(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	verifier := r.URL.Query().Get("oauth_verifier")
	reqTok := new(flickr.RequestToken)
	err := Load("flickr", "request_token", reqTok, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := NewFlickrClient(c)

	accessToken, err := flickr.GetAccessToken(client, reqTok, verifier)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = Save("flickr", "access_token", accessToken, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func Upload(title string, tags []string, isPublic, isFamily, isFriend bool, readCloser io.ReadCloser, c context.Context) error {
	access_token := new(flickr.OAuthToken)
	err := Load("flickr", "access_token", access_token, c)

	if err != nil {
		return err
	}

	client := NewFlickrClient(c)
	client.OAuthToken = access_token.OAuthToken
	client.OAuthTokenSecret = access_token.OAuthTokenSecret

	params := flickr.NewUploadParams()
	params.IsFamily = isFamily
	params.IsFriend = isFriend
	params.IsPublic = isPublic
	params.Tags = tags

	response, err := flickr.UploadReader(client, readCloser, title, params)

	if err != nil {
		return err
	}

	if response.HasErrors() {
		log.Infof(c, "Failed to upload photo %v: %v", title, response)
		return errors.New(response.ErrorMsg())
	} else {
		log.Infof(c, "Uploaded photo %v %v as %v", title, tags, response.ID)
	}

	return nil
}