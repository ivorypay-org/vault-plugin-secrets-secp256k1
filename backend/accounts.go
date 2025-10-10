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
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	// InvalidAddress intends to prevent empty address_to
	InvalidAddress string = "InvalidAddress"
)

// Account is an Ethereum account
type Account struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// KeyPair is a structure to hold the hex encoded private and public keys
type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func paths(b *backend) []*framework.Path {
	return []*framework.Path{
		pathCreateKey(b),
		pathGetKey(b),
		pathSignDigest(b),
	}
}

// func (b *backend) listKeys(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
// 	queryString := data.Get("id").(string)
// 	if queryString != "" {
// 		// If an id is provided, we will return the specific key if it exists
// 		keyPair, err := b.getKey(queryString, req.Storage, ctx)
// 		if err != nil {
// 			b.Logger().Error("Failed to retrieve the key by id", "id", queryString, "error", err)
// 			return nil, err
// 		}
// 		if keyPair == nil {
// 			return logical.ErrorResponse("Key not found"), nil
// 		}
// 		return &logical.Response{
// 			Data: map[string]interface{}{
// 				"id":         queryString,
// 				"privateKey": keyPair.PrivateKey,
// 				"publicKey":  keyPair.PublicKey,
// 			},
// 		}, nil
// 	}
// 	// If no id is provided, we will list all keys
// 	keys, err := req.Storage.List(ctx, "secp256k1/keys/")
// 	if err != nil {
// 		b.Logger().Error("Failed to retrieve the list of keys", "error", err)
// 		return nil, err
// 	}

// 	return logical.ListResponse(keys), nil
// }

func (b *backend) createSecp256k1(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string) // id for this keypair, a good id is a hash(<unique values>)
	var privateKey *ecdsa.PrivateKey

	if id == "" {
		return logical.ErrorResponse("a unique identifier for the keypair is required as a parameter called id"), nil
	}

	// check if the key already exists
	key, err := getKey(id, req.Storage, ctx, b.Logger())
	if err != nil {
		b.Logger().Error("Failed to retrieve the key by id", "id", id, "error", err)
		return logical.ErrorResponse(fmt.Errorf("%v", err).Error()), nil
	}
	if key != nil {
		b.Logger().Info("Key already exists", "id", id)
		return &logical.Response{
			Data: map[string]any{
				"pubKey": key.PublicKey,
			},
		}, nil
	}
	privateKeyHex := data.Get("private_key").(string)
	if privateKeyHex != "" {
		privateKey, err = crypto.HexToECDSA(privateKeyHex[2:])
		if err != nil {
			b.Logger().Error("Failed to parse the provided private key", "error", err)
			return logical.ErrorResponse("failed to parse the provided private key"), nil
		}
	} else {
		privateKey, _ = crypto.GenerateKey()
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyString := hexutil.Encode(privateKeyBytes)[2:]

	defer ZeroKey(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyString := hexutil.Encode(publicKeyBytes)[2:]

	keypair := &KeyPair{
		PrivateKey: privateKeyString,
		PublicKey:  publicKeyString,
	}

	accountPath := fmt.Sprintf("keys/%s", id)

	entry, _ := logical.StorageEntryJSON(accountPath, keypair)
	err = req.Storage.Put(ctx, entry)
	if err != nil {
		b.Logger().Error("failed to save the new keypair to storage", "error", err)
		return logical.ErrorResponse(fmt.Errorf("%v", err).Error()), nil
	}

	return &logical.Response{
		Data: map[string]any{
			"pubKey": publicKeyString,
		},
	}, nil
}

// func (b *backend) exportAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
// 	address := data.Get("name").(string)
// 	b.Logger().Info("Retrieving account for address", "address", address)
// 	account, err := b.getKey(address, req.Storage, ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if account == nil {
// 		return nil, fmt.Errorf("Account does not exist")
// 	}

// 	return &logical.Response{
// 		Data: map[string]interface{}{
// 			"address":    account.PublicKey,
// 			"privateKey": account.PrivateKey,
// 		},
// 	}, nil
// }

// func (b *backend) deleteAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
// 	address := data.Get("name").(string)
// 	account, err := b.getKey(address, req.Storage, ctx)
// 	if err != nil {
// 		b.Logger().Error("Failed to retrieve the account by address", "address", address, "error", err)
// 		return nil, err
// 	}
// 	if account == nil {
// 		return nil, nil
// 	}
// 	if err := req.Storage.Delete(ctx, fmt.Sprintf("accounts/%s", account.PublicKey)); err != nil {
// 		b.Logger().Error("Failed to delete the account from storage", "address", address, "error", err)
// 		return nil, err
// 	}
// 	return nil, nil
// }

func (b *backend) ListKeys(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string)
	if id != "" {
		key, err := getKey(id, req.Storage, ctx, b.Logger())

		if err != nil || key == nil {
			return logical.ErrorResponse(fmt.Errorf("%v", err).Error()), nil
		}
		return &logical.Response{Data: map[string]any{
			"pubKey": key.PublicKey,
		}}, nil
	}

	keys, err := req.Storage.List(ctx, "keys/")
	if err != nil {
		b.Logger().Error("Failed to retrieve the list of accounts", "error", err)
		return logical.ErrorResponse(fmt.Errorf("%v", err).Error()), nil
	}

	return logical.ListResponse(keys), nil
}

func getKey(id string, storage logical.Storage, ctx context.Context, logger log.Logger) (*KeyPair, error) {
	if id == "" {
		return nil, fmt.Errorf("a unique identifier for the keypair is required")
	}
	path := fmt.Sprintf("keys/%s", id)
	entry, err := storage.Get(ctx, path)
	if err != nil || entry == nil {
		logger.Error("Failed to retrieve the key by id", "path", path, "error", err)
		return nil, err
	}
	var keyPair KeyPair
	if err := entry.DecodeJSON(&keyPair); err != nil {
		logger.Error("Failed to decode the key entry", "path", path, "error", err)
		return nil, fmt.Errorf("failed to decode the key entry")
	}
	return &keyPair, nil
}

func (b *backend) sign(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string)
	digestHex := data.Get("data").(string)

	if len(digestHex) != 66 || digestHex[:2] != "0x" {
		return logical.ErrorResponse("digest must be 0x-prefixed 32-byte hex"), nil
	}
	digest, err := hexutil.Decode(digestHex)
	if err != nil || len(digest) != 32 {
		return logical.ErrorResponse("invalid digest"), nil
	}

	key, err := getKey(id, req.Storage, ctx, b.Logger())
	if err != nil || key == nil {
		return logical.ErrorResponse("key not found"), nil
	}

	priv, err := crypto.HexToECDSA(key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("bad private key")
	}
	defer ZeroKey(priv)

	// Use Ethereum-compatible signing which returns [R || S || V]
	signatureBytes, err := crypto.Sign(digest, priv)
	if err != nil {
		return nil, fmt.Errorf("sign failed: %w", err)
	}

	rBytes := signatureBytes[0:32]
	sBytes := signatureBytes[32:64]
	vBytes := signatureBytes[64]

	return &logical.Response{
		Data: map[string]any{
			"r":         "0x" + hex.EncodeToString(rBytes),
			"s":         "0x" + hex.EncodeToString(sBytes),
			"v":         vBytes,
			"signature": "0x" + hex.EncodeToString(signatureBytes),
		},
	}, nil
}

func ZeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}
