// Copyright © 2020 Kaleido
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"context"
	"errors"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/stretchr/testify/assert"
)

func getBackend(t *testing.T) (logical.Backend, logical.Storage) {
	config := &logical.BackendConfig{
		Logger:      logging.NewVaultLogger(log.Trace),
		System:      &logical.StaticSystemView{},
		StorageView: &logical.InmemStorage{},
		BackendUUID: "test",
	}

	b, err := Factory(context.Background(), config)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	// Wait for the upgrade to finish
	time.Sleep(time.Second)

	return b, config.StorageView
}

type StorageMock struct {
	switches []int
}

func (s StorageMock) List(c context.Context, path string) ([]string, error) {
	if s.switches[0] == 1 {
		return []string{"key1", "key2"}, nil
	} else {
		return nil, errors.New("Bang for List!")
	}
}
func (s StorageMock) Get(c context.Context, path string) (*logical.StorageEntry, error) {
	if s.switches[1] == 2 {
		var entry logical.StorageEntry
		return &entry, nil
	} else if s.switches[1] == 1 {
		return nil, nil
	} else {
		return nil, errors.New("Bang for Get!")
	}
}
func (s StorageMock) Put(c context.Context, se *logical.StorageEntry) error {
	return errors.New("Bang for Put!")
}
func (s StorageMock) Delete(c context.Context, path string) error {
	return errors.New("Bang for Delete!")
}

func newStorageMock() StorageMock {
	var sm StorageMock
	sm.switches = []int{0, 0, 0, 0}
	return sm
}

func TestAccounts(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)

	// create key1
	req := logical.TestRequest(t, logical.UpdateOperation, "keys/create")
	req.Data = map[string]any{
		"id": "key1",
	}

	res, err := b.HandleRequest(context.Background(), req)
	storage := req.Storage
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	address1, ok := res.Data["pubKey"].(string)
	assert.True(ok)
	assert.NotEmpty(address1)


	// read account by address
	req = logical.TestRequest(t, logical.ReadOperation, "keys/key1")
	req.Storage = storage
	res, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pubKey, ok := res.Data["pubKey"].(string)
	assert.True(ok)
	assert.Equal(address1, pubKey)

	// read all keys
	req = logical.TestRequest(t, logical.ReadOperation, "keys/")
	req.Storage = storage
	res, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	keys, ok := res.Data["keys"].([]string)
	assert.True(ok)
	assert.Len(keys, 1)
	assert.Equal("key1", keys[0])

	// read non-existent key
	req = logical.TestRequest(t, logical.ReadOperation, "keys/key2")
	req.Storage = storage
	res, err = b.HandleRequest(context.Background(), req)
	assert.Nil(res)
	assert.NotNil(err)
}