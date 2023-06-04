# HashiCorp Vault

Using Vault to Store Slack Keys

## Requirements
The package `vault-client-go` was used here, so you just need to define two variables:
- `VAULT_ADDR`
- `VAULT_TOKEN`

For more information, see [here](https://github.com/hashicorp/vault-client-go).

## Using the utility with this storage method

> With configuration file

```yaml
storage: vault
vault:
  secret_name: secret
  secret_path: mount/path/foo/bar
```

> With environment variables
```shell
ROTATOR_STORAGE=vault
ROTATOR_VAULT_SECRET_NAME=secret
ROTATOR_VAULT_SECRET_PATH=mount/path/foo/bar
```
