package d2f_transfer

import (
	"net/http"
	"encoding/json"
	"google.golang.org/appengine"
)

type TransferRequest struct {
	Title   string
	Tags    []string
	IsPublic, IsFamily, IsFriend bool
	Validations *Validations
	RecentChange                 bool
}

type TransferResponse struct {
	Id string
}

func Transfer(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	transferRequest := new(TransferRequest)

	json.NewDecoder(r.Body).Decode(transferRequest)

	readCloser, _, err := OpenStreamForFile(transferRequest.Title, transferRequest.Validations, transferRequest.RecentChange, c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := Upload(
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

	w.WriteHeader(http.StatusCreated)

	responseData, _ := json.Marshal(&TransferResponse{Id: response.ID})

	w.Write(responseData)
}