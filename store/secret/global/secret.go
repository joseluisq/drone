// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

// +build !oss

package global

import (
	"context"

	"github.com/joseluisq/drone/core"
	"github.com/joseluisq/drone/store/shared/db"
	"github.com/joseluisq/drone/store/shared/encrypt"
)

// New returns a new global Secret database store.
func New(db *db.DB, enc encrypt.Encrypter) core.GlobalSecretStore {
	return &secretStore{
		db:  db,
		enc: enc,
	}
}

type secretStore struct {
	db  *db.DB
	enc encrypt.Encrypter
}

func (s *secretStore) List(ctx context.Context, namespace string) ([]*core.Secret, error) {
	var out []*core.Secret
	err := s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		params := map[string]interface{}{"secret_namespace": namespace}
		stmt, args, err := binder.BindNamed(queryNamespace, params)
		if err != nil {
			return err
		}
		rows, err := queryer.Query(stmt, args...)
		if err != nil {
			return err
		}
		out, err = scanRows(s.enc, rows)
		return err
	})
	return out, err
}

func (s *secretStore) ListAll(ctx context.Context) ([]*core.Secret, error) {
	var out []*core.Secret
	err := s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		rows, err := queryer.Query(queryAll)
		if err != nil {
			return err
		}
		out, err = scanRows(s.enc, rows)
		return err
	})
	return out, err
}

func (s *secretStore) Find(ctx context.Context, id int64) (*core.Secret, error) {
	out := &core.Secret{ID: id}
	err := s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		params, err := toParams(s.enc, out)
		if err != nil {
			return err
		}
		query, args, err := binder.BindNamed(queryKey, params)
		if err != nil {
			return err
		}
		row := queryer.QueryRow(query, args...)
		return scanRow(s.enc, row, out)
	})
	return out, err
}

func (s *secretStore) FindName(ctx context.Context, namespace, name string) (*core.Secret, error) {
	out := &core.Secret{Name: name, Namespace: namespace}
	err := s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		params, err := toParams(s.enc, out)
		if err != nil {
			return err
		}
		query, args, err := binder.BindNamed(queryName, params)
		if err != nil {
			return err
		}
		row := queryer.QueryRow(query, args...)
		return scanRow(s.enc, row, out)
	})
	return out, err
}

func (s *secretStore) Create(ctx context.Context, secret *core.Secret) error {
	if s.db.Driver() == db.Postgres {
		return s.createPostgres(ctx, secret)
	}
	return s.create(ctx, secret)
}

func (s *secretStore) create(ctx context.Context, secret *core.Secret) error {
	return s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params, err := toParams(s.enc, secret)
		if err != nil {
			return err
		}
		stmt, args, err := binder.BindNamed(stmtInsert, params)
		if err != nil {
			return err
		}
		res, err := execer.Exec(stmt, args...)
		if err != nil {
			return err
		}
		secret.ID, err = res.LastInsertId()
		return err
	})
}

func (s *secretStore) createPostgres(ctx context.Context, secret *core.Secret) error {
	return s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params, err := toParams(s.enc, secret)
		if err != nil {
			return err
		}
		stmt, args, err := binder.BindNamed(stmtInsertPg, params)
		if err != nil {
			return err
		}
		return execer.QueryRow(stmt, args...).Scan(&secret.ID)
	})
}

func (s *secretStore) Update(ctx context.Context, secret *core.Secret) error {
	return s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params, err := toParams(s.enc, secret)
		if err != nil {
			return err
		}
		stmt, args, err := binder.BindNamed(stmtUpdate, params)
		if err != nil {
			return err
		}
		_, err = execer.Exec(stmt, args...)
		return err
	})
}

func (s *secretStore) Delete(ctx context.Context, secret *core.Secret) error {
	return s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params, err := toParams(s.enc, secret)
		if err != nil {
			return err
		}
		stmt, args, err := binder.BindNamed(stmtDelete, params)
		if err != nil {
			return err
		}
		_, err = execer.Exec(stmt, args...)
		return err
	})
}

const queryBase = `
SELECT
 secret_id
,secret_namespace
,secret_name
,secret_type
,secret_data
,secret_pull_request
,secret_pull_request_push
`

const queryKey = queryBase + `
FROM orgsecrets
WHERE secret_id = :secret_id
LIMIT 1
`

const queryAll = queryBase + `
FROM orgsecrets
ORDER BY secret_name
`

const queryName = queryBase + `
FROM orgsecrets
WHERE secret_name = :secret_name
  AND secret_namespace = :secret_namespace
LIMIT 1
`

const queryNamespace = queryBase + `
FROM orgsecrets
WHERE secret_namespace = :secret_namespace
ORDER BY secret_name
`

const stmtUpdate = `
UPDATE orgsecrets SET
 secret_data = :secret_data
,secret_pull_request = :secret_pull_request
,secret_pull_request_push = :secret_pull_request_push
WHERE secret_id = :secret_id
`

const stmtDelete = `
DELETE FROM orgsecrets
WHERE secret_id = :secret_id
`

const stmtInsert = `
INSERT INTO orgsecrets (
 secret_namespace
,secret_name
,secret_type
,secret_data
,secret_pull_request
,secret_pull_request_push
) VALUES (
 :secret_namespace
,:secret_name
,:secret_type
,:secret_data
,:secret_pull_request
,:secret_pull_request_push
)
`

const stmtInsertPg = stmtInsert + `
RETURNING secret_id
`
