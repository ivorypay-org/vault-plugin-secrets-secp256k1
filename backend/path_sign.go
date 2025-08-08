package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathSignDigest(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "keys/" + framework.GenericNameRegex("id") + "/sign-digest",
		Fields: map[string]*framework.FieldSchema{
			"id": {
				Type:        framework.TypeString,
				Description: "Key ID",
				Required:    true,
			},
			"digest": {
				Type:        framework.TypeString,
				Description: "32-byte Keccak-256 digest (0x...)",
				Required:    true,
			},
			// Optional knobs if you ever want them:
			"return": {
				Type:        framework.TypeString,
				Description: "one of: compact (default), components",
				Default:     "compact",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			// support both POST and PUT so `vault write` works
			logical.CreateOperation: b.sign,
			logical.UpdateOperation: b.sign,
		},
		HelpSynopsis:    "Sign a 32-byte digest with a secp256k1 key",
		HelpDescription: "Returns r, s, and recovery id (yParity). Caller assembles the final transaction.",
	}
}
