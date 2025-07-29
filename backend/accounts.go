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
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
		pathListKeys(b),
	}
}

func (b *backend) listKeys(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	queryString := data.Get("id").(string)
	if queryString != "" {
		// If an id is provided, we will return the specific key if it exists
		keyPair, err := b.getKey(queryString, req.Storage, ctx)
		if err != nil {
			b.Logger().Error("Failed to retrieve the key by id", "id", queryString, "error", err)
			return nil, err
		}
		if keyPair == nil {
			return logical.ErrorResponse("Key not found"), nil
		}
		return &logical.Response{
			Data: map[string]interface{}{
				"id":         queryString,
				"privateKey": keyPair.PrivateKey,
				"publicKey":  keyPair.PublicKey,
			},
		}, nil
	}
	// If no id is provided, we will list all keys
	keys, err := req.Storage.List(ctx, "secp256k1/keys/")
	if err != nil {
		b.Logger().Error("Failed to retrieve the list of keys", "error", err)
		return nil, err
	}

	return logical.ListResponse(keys), nil
}

func (b *backend) createSecp256k1(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	id := data.Get("id").(string) // id for this keypair, a good id is a hash(<unique values>)
	var privateKey *ecdsa.PrivateKey

	if id == "" {
		return nil, fmt.Errorf("a unique identifier for the keypair is required as a parameter called id")
	}

	privateKey, _ = crypto.GenerateKey()
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

	accountPath := fmt.Sprintf("secp256k1/keys/%s", id)
	accountBz, _ := json.Marshal(keypair)

	entry, _ := logical.StorageEntryJSON(accountPath, accountBz)
	err := req.Storage.Put(ctx, entry)
	if err != nil {
		b.Logger().Error("failed to save the new keypair to storage", "error", err)
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"pubKey": publicKeyString,
		},
	}, nil
}

func (b *backend) exportAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	address := data.Get("name").(string)
	b.Logger().Info("Retrieving account for address", "address", address)
	account, err := b.getKey(address, req.Storage, ctx)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("Account does not exist")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"address":    account.PublicKey,
			"privateKey": account.PrivateKey,
		},
	}, nil
}

func (b *backend) deleteAccount(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	address := data.Get("name").(string)
	account, err := b.getKey(address, req.Storage, ctx)
	if err != nil {
		b.Logger().Error("Failed to retrieve the account by address", "address", address, "error", err)
		return nil, err
	}
	if account == nil {
		return nil, nil
	}
	if err := req.Storage.Delete(ctx, fmt.Sprintf("accounts/%s", account.PublicKey)); err != nil {
		b.Logger().Error("Failed to delete the account from storage", "address", address, "error", err)
		return nil, err
	}
	return nil, nil
}

func (b *backend) getKey(id string, storage logical.Storage, ctx context.Context) (*KeyPair, error) {
	if id == "" {
		return nil, fmt.Errorf("a unique identifier for the keypair is required")
	}
	path := fmt.Sprintf("secp256k1/keys/%s", id)
	entry, err := storage.Get(ctx, path)
	if err != nil {
		b.Logger().Error("Failed to retrieve the key by id", "path", path, "error", err)
		return nil, err
	}
	if entry == nil {
		// could not find the corresponding key for the id
		return nil, nil
	}
	var keyPair KeyPair
	if err := entry.DecodeJSON(&keyPair); err != nil {
		b.Logger().Error("Failed to decode the key entry", "path", path, "error", err)
		return nil, fmt.Errorf("failed to decode the key entry")
	}
	return &keyPair, nil
}

func (b *backend) signTx(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	from := data.Get("name").(string)

	var txDataToSign []byte
	dataInput := data.Get("data").(string)
	// some client such as go-ethereum uses "input" instead of "data"
	if dataInput == "" {
		dataInput = data.Get("input").(string)
	}
	if len(dataInput) > 2 && dataInput[0:2] != "0x" {
		dataInput = "0x" + dataInput
	}

	txDataToSign, err := hexutil.Decode(dataInput)
	if err != nil {
		b.Logger().Error("Failed to decode payload for the 'data' field", "error", err)
		return nil, err
	}

	account, err := b.getKey(from, req.Storage, ctx)
	if err != nil {
		b.Logger().Error("failed to retrieve the signing account", "address", from, "error", err)
		return nil, fmt.Errorf("error retrieving signing account %s", from)
	}
	if account == nil {
		return nil, fmt.Errorf("signing account %s does not exist", from)
	}
	amount := ValidNumber(data.Get("value").(string))
	if amount == nil {
		b.Logger().Error("invalid amount for the 'value' field", "value", data.Get("value").(string))
		return nil, fmt.Errorf("invalid amount for the 'value' field")
	}

	rawAddressTo := data.Get("to").(string)

	chainId := ValidNumber(data.Get("chainId").(string))
	if chainId == nil {
		b.Logger().Error("invalid chainId", "chainId", data.Get("chainId").(string))
		return nil, fmt.Errorf("invalid 'chainId' value")
	}

	gasLimitIn := ValidNumber(data.Get("gas").(string))
	if gasLimitIn == nil {
		b.Logger().Error("invalid gas limit", "gas", data.Get("gas").(string))
		return nil, fmt.Errorf("invalid gas limit")
	}
	gasLimit := gasLimitIn.Uint64()

	gasPrice := ValidNumber(data.Get("gasPrice").(string))

	privateKey, err := crypto.HexToECDSA(account.PrivateKey)
	if err != nil {
		b.Logger().Error("error reconstructing private key from retrieved hex", "error", err)
		return nil, fmt.Errorf("error reconstructing private key from retrieved hex")
	}
	defer ZeroKey(privateKey)

	nonceIn := ValidNumber(data.Get("nonce").(string))
	nonce := nonceIn.Uint64()

	var tx *types.Transaction
	if rawAddressTo == "" {
		tx = types.NewContractCreation(nonce, amount, gasLimit, gasPrice, txDataToSign)
	} else {
		toAddress := common.HexToAddress(rawAddressTo)
		tx = types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, txDataToSign)
	}
	var signer types.Signer
	if big.NewInt(0).Cmp(chainId) == 0 {
		signer = types.HomesteadSigner{}
	} else {
		signer = types.NewEIP155Signer(chainId)
	}
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		b.Logger().Error("Failed to sign the transaction object", "error", err)
		return nil, err
	}

	var signedTxBuff bytes.Buffer
	signedTx.EncodeRLP(&signedTxBuff)

	return &logical.Response{
		Data: map[string]interface{}{
			"transaction_hash":   signedTx.Hash().Hex(),
			"signed_transaction": hexutil.Encode(signedTxBuff.Bytes()),
		},
	}, nil
}

func ValidNumber(input string) *big.Int {
	if input == "" {
		return big.NewInt(0)
	}
	matched, err := regexp.MatchString("([0-9])", input)
	if !matched || err != nil {
		return nil
	}
	amount := math.MustParseBig256(input)
	return amount.Abs(amount)
}

func ZeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}
