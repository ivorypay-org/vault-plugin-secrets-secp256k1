package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCreateKey(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/create$",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.CreateOperation: b.createSecp256k1,
			logical.UpdateOperation: b.createSecp256k1,
		},
		ExistenceCheck: b.pathExistenceCheck,
		HelpSynopsis:    "create a secp256k1 key",
		HelpDescription: "Post to this endpoint to create a secp256k1 key. The path end is the hash where the key is stored",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "The ID of the key to create. This field is required.",
				Required:    true,
			},
			"private_key": {
				Type:        framework.TypeString,
				Description: "Optional hex-encoded private key. If not provided, a new key will be generated.",
				Required:    false,
			},
		},
	}
}

func pathGetKey(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys(/" + framework.GenericNameRegex("id") + ")?/?$",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.ListKeys,
			logical.ListOperation: b.ListKeys,
		},
		HelpSynopsis:    "List or retrieve keys",
		HelpDescription: "Lists all keys or retrieves a specific key by ID.",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "The ID of the key to retrieve. If omitted, all keys will be listed.",
				Required:    false,
			},
		},
	}
}