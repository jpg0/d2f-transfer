package d2f_transfer

import (
	"google.golang.org/appengine/user"
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func CanConfigure (h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		c := appengine.NewContext(r)

		if !user.IsAdmin(c) {
			url, err := user.LoginURL(c, r.RequestURI)

			if err != nil { //shouldn't happen
				panic(err)
			}

			http.Redirect(w, r, url, http.StatusUnauthorized)
			return
		}

		h(w, r)
	}
}

func CanTransfer(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type Password struct {
			Value string
		}


		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		_, password, _ := r.BasicAuth()

		if password == "" {
			http.Error(w, "Not authorized", 401)
			return
		}

		c := appengine.NewContext(r)
		k := datastore.NewKey(c, "Password", "password", 0, nil)
		p := new(Password)
		if err := datastore.Get(c, k, p); err != nil {
			// If password is not set, seed with whatever password was passed in.
			// See: http://golang.org/misc/dashboard/app/build/key.go
			dp := Password{
				Value: password,
			}
			if _, err := datastore.Put(c, k, &dp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			h(w, r)
		} else if p.Value == password { //success
			h(w, r)
		} else {
			http.Error(w, "Not authorized", 401)
		}
		return
	}
}