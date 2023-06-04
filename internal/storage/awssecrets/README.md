# AWS Secrets

Using AWS Secrets to Store Slack Keys

## Requirements
The package `aws-sdk-go-v2` was used here, so you just need to define two variables:
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`

For more information, see [here](https://github.com/aws/aws-sdk-go-v2).

## Using the utility with this storage method

> With configuration file

```yaml
storage: awssecrets
awssecrets:
  secret_name: secret
```

> With environment variables
```shell
ROTATOR_STORAGE=awssecrets
ROTATOR_AWSSECRETS_SECRET_NAME=secret
```
