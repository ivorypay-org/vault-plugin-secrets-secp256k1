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
		ExistenceCheck:  b.pathExistenceCheck,
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
		// Pattern explanation:
		// - "keys/" — base path
		// - "(?P<product>[\w\-.]+)" — capture the product segment
		// - "(/(?P<id>[\w\-.]+))?" — optional key ID after the product
		// - "/?$" — optional trailing slash
		Pattern: `keys/(?P<product>[\w\-.]+)(/(?P<id>[\w\-.]+))?/?$`,

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.ListKeys,
			logical.ListOperation: b.ListKeys,
		},

		HelpSynopsis: "List or retrieve product-partitioned keys.",
		HelpDescription: `
			This endpoint allows you to list or retrieve keys within a product namespace.

			- GET  /keys/<product>           → Lists all key IDs under that product.
			- GET  /keys/<product>/<id>      → Retrieves the key for that product and ID.
			- LIST /keys/<product>           → Lists all key IDs (Vault LIST mode).
			`,
		Fields: map[string]*framework.FieldSchema{
			"product": {
				Type:        framework.TypeString,
				Description: "Product namespace for the key (e.g. duffle, IvoryHolding).",
				Required:    true,
			},
			"id": {
				Type:        framework.TypeString,
				Description: "ID of the key to retrieve (optional for listing).",
				Required:    false,
			},
		},
	}
}
