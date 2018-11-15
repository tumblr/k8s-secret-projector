# Hacking

Make sure you have a working go env first.

```bash
make && ./bin/k8s-secret-projector -h
```

## Linux Native

```bash
$ GOOS=linux GOARCH=amd64 make
$ file bin/k8s-secret-projector
bin/k8s-forward-zone-generator: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, not stripped
```

## Docker

```bash
$ ./ci-cd/build.sh
...
Successfully built 8ce7a5f5542c
Successfully tagged tumblr/k8s-secret-projector:v0.1.0-13-g1f564dc
```

NOTE: CI is setup to automatically build images and push them to Docker Hub

## Dependencies

We use go 1.11+ modules for dependencies. See upstream docs at https://github.com/golang/go/wiki/Modules

## Maintainer

Gabe (gabe@tumblr.com)
