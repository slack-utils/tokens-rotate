# Refreshing Slack configuration tokens

Applications in Slack can be managed by code, but this requires a configuration token, the lifetime of which is only 12 hours.
Slack provides two tokens - `access_token` and `refresh_token`, and with refresh_token you can update `access_token`.

For more information about configuration tokens you can see [here](https://api.slack.com/authentication/config-tokens).

This utility can update the `access_token` before it expires and store it in storage.

As a storage, you can choose one of the following:
- [AWS Secrets](internal/storage/awssecrets)
- [Filesystem](internal/storage/fs)
- [Hashicorp Vault](internal/storage/vault)

## Requirements
To run the utility, you need to pass `refresh_token` through environment variables:
- `ROTATOR_REFRESH_TOKEN` - xoxe-1-***

If there are already tokens in the storage and they have expired, tokens from the environment variables will be used and stored in the storage.
