# Filesystem

Storing Slack keys in the filesystem

## Using the utility with this storage method

> With configuration file

```yaml
storage: fs
fs:
  token_file: /path/to/file.json
```

> With environment variables
```shell
ROTATOR_STORAGE=fs
ROTATOR_FS_TOKEN_FILE=/path/to/file.json
```
