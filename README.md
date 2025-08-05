# vault-plugin-secrets-ethsign

[![Build Status](https://travis-ci.org/kaleido-io/vault-plugin-secrets-ethsign.svg?branch=master)](https://travis-ci.org/kaleido-io/vault-plugin-secrets-ethsign)
[![codecov](https://codecov.io/gh/kaleido-io/vault-plugin-secrets-ethsign/branch/master/graph/badge.svg?token=3LlJ7aSeW2)](https://codecov.io/gh/kaleido-io/vault-plugin-secrets-ethsign)

A HashiCorp Vault plugin that supports secp256k1 based signing, with an API interface that turns the vault into a software-based HSM device.

![Overview](/resources/overview.png)

The plugin only exposes the following endpoints to enable the client to generate signing keys for the secp256k1 curve. List existing signing keys by their id, and a `/sign` endpoint for each account. The generated private keys are saved in the vault as a secret. It never gives out the private keys.

## Build
These dependencies are needed:

* go 1.16

To build the binary:
```
make all
```

The output is `ivory-secp256k1`, which is the plugin binary that can be used with HashiCorp Vault.

## Installing the Plugin on HashiCorp Vault server
The plugin must be registered and enabled on the vault server as a secret engine.

### Enabling on a dev mode server
The easiest way to try out the plugin is using a dev mode server to load it.

Download the binary: [https://www.vaultproject.io/downloads/](https://www.vaultproject.io/downloads/)

First copy the build output binary `ivory-secp256k1` to the plugins folder, say `~/.vault.d/vault-plugins/`.
```
./vault server -dev -dev-plugin-dir=/Users/alice/.vault.d/vault_plugins/
```

After the dev server starts, the plugin should have already been registered in the system plugins catalog:
```
$ ./vault login <root token>
$ ./vault read sys/plugins/catalog
Key         Value
---         -----
auth        [alicloud app-id approle aws azure centrify cert cf gcp github jwt kubernetes ldap oci oidc okta pcf radius userpass]
database    [cassandra-database-plugin elasticsearch-database-plugin hana-database-plugin influxdb-database-plugin mongodb-database-plugin mssql-database-plugin mysql-aurora-database-plugin mysql-database-plugin mysql-legacy-database-plugin mysql-rds-database-plugin postgresql-database-plugin]
secret      [ad alicloud aws azure cassandra consul ivory-secp256k1 gcp gcpkms kv mongodb mssql mysql nomad pki postgresql rabbitmq ssh totp transit]
```

Note the `ivory-secp256k1` entry in the secret section. Now it's ready to be enabled:
```
 ./vault secrets enable -path=secp256k1 -description="SECP 256k1" -plugin-name=ivory-secp256k1 plugin
```

To verify the new secret engine based on the plugin has been enabled:
```
$ ./vault secrets list
Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_1f1e372d    per-token private secret storage
ethereum/     ethsign      ethsign_d9f104c7      Ethereum Wallet
identity/     identity     identity_382e2000     identity store
secret/       kv           kv_32f5a684           key/value secret storage
sys/          system       system_21e0c7c7       system endpoints used for control, policy and debugging
```

### Enabling on a non-dev mode server
Setting up a non-dev mode server is beyond the scope of this README, as this is a very sensitive IT operation. But a simple procedure can be found in [the wiki page](https://github.com/kaleido-io/vault-plugin-secrets-ethsign/wiki/Setting-Up-A-Local-HashiCorp-Vault-Server).

Before enabling the plugin on the server, it must first be registered.

First copy the binary to the plugin folder for the server (consult the configuration file for the plugin folder location). Then calculate a SHA256 hash for the binary.
```
shasum -a 256 ./ivory-secp256k1
```

Use the hash to register the plugin with vault:
```
 ./vault write sys/plugins/catalog/eth-hsm sha_256=$SHA command="ivory-secp256k1"
```
> If the target vault server is enabled for TLS, and is using a self-signed certificate or other non-verifiable TLS certificate, then the command value needs to contain the switch to turn off TLS verify: `command="ivory-secp256k1 -tls-skip-verify"`

Once registered, just like in dev mode, it's ready to be enabled as a secret engine:
```
 ./vault secrets enable -path=ethereum -description="Eth Signing Wallet" -plugin-name=ivory-secp256k1 plugin
```

## Interacting with the ivory-secp256k1 Plugin
The plugin does not interact with the target blockchain. It has very simple responsibilities: sign transactions f blockchain.
It does not validate the transactions, nor does it check the balances of the accounts. It is up to the client to ensure that the transactions are valid and that the accounts have sufficient balance to pay for the gas.
### Creating A New Signing Key
Create a new key in the vault by POSTing to the `/keys` endpoint.

Using the REST API:
```
$ curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{}' http://localhost:8200/v1/secp256k1/keys |jq

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0xb579cbf259a8d36b22f2799eeeae5f3553b11eb7"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault write -force secp256k1/keys/

Key       Value
---       -----
pubKey    0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10
```

### List Existing Keys
The list command only returns the addresses of the signing accounts. To return the private keys, use the `/export/keys/:id` endpoint.

Using the REST API:
```
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp256k1/keys?list=true |jq

{
  "request_id": "56c31ef5-9757-1ff4-354e-3b18ecd8ea77",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "keys": [
      "0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10",
    ]
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault list secp256k1/keys

Keys
----

```

### Reading Individual Keys
Inspect the key using the address. Only the address of the signing key is returned. To return the private key, use the `/export/keys/:address` endpoint.

Using the REST API:
```
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp256k1/keys/0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10 |jq

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "pubKey": "0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10",
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault read secp256k1/keys/0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10

Key        Value
---        -----
pubKey     0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10
```

### Export A Key
To export a key, you need to use the `/export/keys/:address` endpoint.
This will return the private key in addition to the address. This is useful for importing the key into another wallet or signing service.
> **Warning**: Exporting the private key is a sensitive operation. Make sure you understand the implications of exporting a private key. Once exported, the key can be used to sign transactions without the vault plugin. It is recommended to only export keys that are not used for signing transactions in production environments, or to export keys that are used for testing purposes only.
You can also export the account by returning the private key.

Using the REST API:
```
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp256k1/export/keys/0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10 |jq

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0xb579cbf259a8d36b22f2799eeeae5f3553b11eb7",
    "privateKey": "ec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault read secp256k1/export/keys/0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10

Key           Value
---           -----
pubKey       0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10
privateKey    ec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2
```

### Sign A Transaction
Use one of the accounts to sign a transaction.

Using the REST API:
```
$  curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp256k1/keys/0457a4d0a822b053b2ed158b877f1613380157e5ea213e82fd8abe9a2b0515a356896c6aae1de2f79b321d7155d4c1bf5bcfcb9f49e0c17fc149db600310596a10/sign -d '{"data":"0x60fe47b10000000000000000000000000000000000000000000000000000000000000014","gas":30791,"gasPrice":0,"nonce":"0x0","to":"0xca0fe7354981aeb9d051e2f709055eb50b774087"}' |jq

{
  "request_id": "4b68c813-eda9-e3c7-4651-e9dbc526bf47",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "signed_transaction": "0xf888808083015f9094b401069f06a24155774bf8a0f6654ea299c8f68780a460fe47b10000000000000000000000000000000000000000000000000000000000000014840ea23e3fa088f4f5505f6f1da6c9a543863d5c7537e0dfc58618dbf34517c80875283d1e07a0583ecdc23ba3333a3f25611fffe0ec7fb585e9b9af93941f6e3ef8c8ef410698",
    "transaction_hash": "0x7ac47960a9398ae73b994c46fcb8834068195a2d3468c40a1eaad7ed4a15e68e"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

To sign a contract deploy, simply skip the `to` parameter in the JSON payload.

To use EIP155 signer, instead of Homestead signer, pass in `chainId` in the JSON payload.

The `signed_transaction` value in the response is already RLP encoded and can be submitted to an Ethereum blockchain directly.

## Access Policies
The plugin's endpoint paths are designed such that admin-level access policies vs. user-level access policies can be easily separated.

### Sample User Level Policy:
Use the following policy to assign to a regular user level access token, with the abilities to list keys, read individual keys and sign transactions.

```
/*
 * Ability to list existing keys ("list")
 */
path "ethereum/accounts" {
  capabilities = ["list"]
}
/*
 * Ability to retrieve individual keys ("read"), sign transactions ("create")
 */
path "ethereum/accounts/*" {
  capabilities = ["create", "read"]
}
```

### Sample Admin Level Policy:
Use the following policy to assign to a admin level access token, with the full ability to create keys, import existing private keys, export private keys, read/delete individual keys, and sign transactions.

```
/*
 * Ability to create key ("update") and list existing keys ("list")
 */
path "ethereum/accounts" {
  capabilities = ["update", "list"]
}
/*
 * Ability to retrieve individual keys ("read"), sign transactions ("create") and delete keys ("delete")
 */
path "ethereum/accounts/*" {
  capabilities = ["create", "read", "delete"]
}
/*
 * Ability to export private keys ("read")
 */
path "ethereum/export/accounts/*" {
  capabilities = ["read"]
}
```
