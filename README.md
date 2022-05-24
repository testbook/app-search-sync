## Realtime sync from mongo to [Elastic App Search](https://www.elastic.co/app-search/)

### TODO

 - [ ] Validation check when last sync mongo token isn't present in [mongo oplog](https://www.mongodb.com/docs/manual/core/replica-set-oplog/)

 - [ ] Logrotate implementation (possibly via independent process)

 - [ ] Support for multiple mongo connections

### Usage

 - Pass config path while running the binary
    - ```bash
        go run . -f {configFilePath}.toml
        ```

