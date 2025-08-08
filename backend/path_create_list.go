package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCreateKey(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/?$",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.CreateOperation: b.createSecp256k1,
			logical.UpdateOperation: b.createSecp256k1,
		},
		HelpSynopsis:    "create a secp256k1 key",
		HelpDescription: "Post to this endpoint to create a secp256k1 key. The path end is the hash where the key is stored",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "The ID of the key to create. This field is required.",
				Required:    true,
			},
		},
	}
}

func pathGetKey(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/" + framework.GenericNameRegex("id"),
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.ListKeys,
		},
			HelpSynopsis:    "List keys by prefix",
		HelpDescription: "Lists keys under a given prefix below keys/.",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "The ID of the key to retrieve. If not provided, all keys will be listed.",
				Required:    true,
			},
		},
	}
}

func pathListKeys(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/?$",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.ListKeys,
		},
		HelpSynopsis:    "List all keys",
		HelpDescription: "Lists all keys stored under the keys/ prefix.",
	}
}
