package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCreateKey(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/?$",
		Callbacks: map[logical.Operation]framework.OperationFunc{
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

func pathListKeys(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/" + framework.GenericNameRegex("id"),
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.ListKeys,
		},
		HelpSynopsis:    "List all or a specific secp256k1 key created or managed by the plugin backend.",
		HelpDescription: "Use this endpoint to retrieve a list of all secp256k1 keys created or managed by the plugin backend. If no keys are found, an empty list will be returned. If an id is provided, it will return the specific key if it exists.",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "The ID of the key to retrieve. If not provided, all keys will be listed.",
				Required:    false,
				Query:       true,
			},
		},
	}
}
