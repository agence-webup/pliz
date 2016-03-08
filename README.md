# Pliz

## Requirements

- Go 1.6 (for vendor support)
- glide (https://github.com/Masterminds/glide)
- _Optional_: gox (https://github.com/mitchellh/gox) for cross compilation

## Build the project

```bash
$ mkdir -p $GOPATH/src/webup
$ git clone git@bitbucket.org:agencewebup/pliz.git pliz
$ cd $GOPATH/src/webup/pliz
$ glide install
$ go install
```

### Cross compilation

```bash
$ cd $GOPATH/src/webup/pliz
$ gox -osarch="linux/amd64"
```
