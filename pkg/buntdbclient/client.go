// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// buntdbclient implements ClientStore from gopkg.in/oauth2.v3
// using BuntDB (https://github.com/tidwall/buntdb).
package buntdbclient

import (
	"fmt"
	"time"

	"github.com/tidwall/buntdb"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

var (
	// DefaultTTL is the value used as TTL on buntdb.SetOptions
	DefaultTTL time.Duration = 24 * time.Hour
)

func New(path string) (oauth2.ClientStore, error) {
	db, err := buntdb.Open(path)
	if err != nil {
		return nil, err
	}
	return &ClientStore{
		db: db,
	}, nil
}

type ClientStore struct {
	oauth2.ClientStore

	db *buntdb.DB
}

func (cs *ClientStore) Close() error {
	return cs.db.Close()
}

func (cs *ClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	var cli models.Client
	cli.ID = id

	err := cs.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(fmt.Sprintf("%s-secret", id))
		if err != nil {
			return err
		}
		cli.Secret = v

		v, err = tx.Get(fmt.Sprintf("%s-domain", id))
		if err != nil {
			return err
		}
		cli.Domain = v

		v, err = tx.Get(fmt.Sprintf("%s-user-id", id))
		if err != nil {
			return err
		}
		cli.UserID = v
		return nil
	})
	if err != nil {
		var cli models.Client
		return &cli, fmt.Errorf("problem reading %s: %v", id, err)
	}
	return &cli, nil
}

func (cs *ClientStore) Set(id string, cli oauth2.ClientInfo) error {
	if inc := cli.GetID(); id != inc {
		return fmt.Errorf("ClientStore: id's don't match, id=%s and cli=%s", id, inc)
	}

	err := cs.db.Update(func(tx *buntdb.Tx) error {
		opts := &buntdb.SetOptions{
			Expires: DefaultTTL > 0,
			TTL:     DefaultTTL,
		}
		_, _, err := tx.Set(fmt.Sprintf("%s-secret", id), cli.GetSecret(), opts)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(fmt.Sprintf("%s-domain", id), cli.GetDomain(), opts)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(fmt.Sprintf("%s-user-id", id), cli.GetUserID(), opts)
		return err
	})
	if err != nil {
		return fmt.Errorf("problem updating %s: %v", id, err)
	}
	return nil
}
