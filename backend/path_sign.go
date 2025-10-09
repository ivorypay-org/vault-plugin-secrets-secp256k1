package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathSignDigest(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/" + framework.GenericNameRegex("id") + "/sign$",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "Key Id",
				Required:    true,
			},
			"data": {
				Type:        framework.TypeString,
				Description: "32 byte hex-encoded digest to sign",
				Required:    true,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.sign,
		},
		ExistenceCheck: b.pathExistenceCheck,
		HelpSynopsis:    "Sign a 32-byte digest with a secp256k1 key",
		HelpDescription: "Returns r, s, and recovery id (yParity)",
	}
}
