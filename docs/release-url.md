## Release URL configurations

- `https/http`: Download release tar from the URL. Exact version (dev or final) must be specified.

```
releases:
- name: bosh
  version: 88
  url: https://s3.amazonaws.com/bosh-jenkins-artifacts/release/bosh-2619.tgz
```

- `dir+bosh`: Use release directory located on the _host_ FS. 
  Exact dev release version or `latest` must be specified.
  If version is `latest` new dev release will be created via `bosh create release --force`.

```
releases:
- name: bosh
  version: latest
  url: dir+bosh://../../bosh
```

- `dir`: Use release directory located on the _guest_ FS. Exact dev release version must be specified.

```
releases:
- name: bosh
  version: 88+dev.1
  url: dir:///tmp/bosh
```

- `file`: Use release tar located on the _guest_ FS. Exact version (dev or final) must be specified.

```
releases:
- name: bosh
  version: 88
  url: file:///tmp/bosh-2619.tgz
```
