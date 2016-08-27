package store

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"encoding/json"
)

func Save(namespace string, name string, value interface{}, c context.Context) error {
	key := datastore.NewKey(c, namespace, name, 0, nil)
	o, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = datastore.Put(c, key, string(o))
	return err
}

func Load(namespace string, name string, value interface{}, c context.Context) error {
	key := datastore.NewKey(c, namespace, name, 0, nil)
	o := new(string)
	err := datastore.Get(c, key, o)

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(*o), value)
}