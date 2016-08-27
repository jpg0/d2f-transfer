package d2f_transfer

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"encoding/json"
)

type SingleValue struct {
	Value string
}

func Save(namespace string, name string, value interface{}, c context.Context) error {
	key := datastore.NewKey(c, namespace, name, 0, nil)
	o, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = datastore.Put(c, key, &SingleValue{Value:string(o)})
	return err
}

func Load(namespace string, name string, value interface{}, c context.Context) error {
	key := datastore.NewKey(c, namespace, name, 0, nil)
	o := new(SingleValue)
	err := datastore.Get(c, key, o)

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(o.Value), value)
}