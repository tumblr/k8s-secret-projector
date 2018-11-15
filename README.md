# k8s-secret-projector

Create managed Kubernetes Secrets from mapping files and your credentials repos

![GitHub release](https://img.shields.io/github/release/tumblr/k8s-secret-projector.svg) ![Travis (.org)](https://img.shields.io/travis/tumblr/k8s-secret-projector.svg) ![Docker Automated build](https://img.shields.io/docker/automated/tumblr/k8s-secret-projector.svg) ![Docker Build Status](https://img.shields.io/docker/build/tumblr/k8s-secret-projector.svg) ![MicroBadger Size](https://img.shields.io/microbadger/image-size/tumblr/k8s-secret-projector.svg) ![Docker Pulls](https://img.shields.io/docker/pulls/tumblr/k8s-secret-projector.svg) ![Docker Stars](https://img.shields.io/docker/stars/tumblr/k8s-secret-projector.svg) [![Godoc](https://godoc.org/github.com/tumblr/k8s-secret-projector?status.svg)](http://godoc.org/github.com/tumblr/k8s-secret-projector)

# What is this?

At Tumblr, we wanted a way to allow applications to declare their dependencies on secrets (passwords, certificates, etc) without needing to create configurations that are aware the specific secret files. A system like this will allow automation to ensure applications always have the appropriate secrets at runtime, while enabling automated systems (cert refreshers, DB password rotations, etc) to automatically manage and update these credentials, and not require the application to redeploy/restart. Additionally, we wanted a system to limit to scope and access of any application to the minimum set of credentials necessary to run, to minimize a compromize blast radius.

To solve this problem, we created a system to enable developers to declare their application's secret dependencies, without needing access to the secret files in our secure credential stores. The `k8s-secret-projector` is a tool that:

- Has access to your credential repos (typically git repo on disk)
- Reads your `ProjectionManifest` YAMLs (typically in a separate git repo)
- Creates a set of Kubernetes `Secret` manifests by processing `ProjectionManifest` along with your credential repos
- Optionally encrypts those Secrets before they touch disk (if your applications have decrypt-before-use ability for your secrets)

It is meant to operate in a larger system of CI/CD in your organization. The following diagram explains a typical setup:

```
+------------------------+
|                        |                        +---------------------------------------------+
| Secret Manifests Repo  |                        |                                             |
| (projection manifests) |                        |  Typical Workflow for k8s-secret-projector  |
|                        |                        |  use in CI/CD pipelines                     |
+-----+------------------+                        |                                             |
      |                                           +---------------------------------------------+
      | Webhook/similar to trigger
      | CI/CD jobs on merge to master                                                                             +------------------+
      |                                                                                                           |                  |
      |                                                                                                           | Credentials Repo |
      |                              +----------------------+                 +-------------------+               | (staging)        |
+-----v----------------+             |                      |                 |                   +--------------->                  |
|                      |             | k8s-secret-projector |                 | Check out         |               +-------------+----+
| Jenkins/Travis CI/CD +-------------> docker image         +-----------------> Credentials Repos +---------+                   |
|                      |             |                      |                 |                   |         |                   |
+----------------------+             +----------------------+                 +--------+----------+    +----v-------------+     |
                                                                                       |               |                  |     |
                     Job runs in container,                                            |               | Credentials Repo |     |
                     sets up workspace                                                 |               | (production)     |     |
                                                                                       |               |                  |     |
                                                                                       |               +----------+-------+     |
                                                                                       |                          |             |
                                                                                       |    +------------------+  |             |
                                                                                       |    |                  |  |             |
                                                                                       +----> Credentials Repo |  |             |
                                                                                            | (development)    |  |             |
                                                                                            |                  |  |             |
                                                                                            +----------------+-+  |             |
                                                                                                             |    |             |
                                                                                                             |    |             |
                                                                                                             |    |             |
                               +-----------------------+               +----------------------------------+  |    |             |
                               |                       |               |                                  <--+    |             |
                               | Use kubectl to deploy <---------------+ Run k8s-secret-projector         <-------+             |
                               | generated secrets     |               | (generate secret YAML resources) <---------------------+
                               |                       |               |                                  |
                               +-----------------------+               +------------------^---------------+
                                                                                          |
                                                                                          |
                                                                                          | This should be the commit that
                                                                                          | triggered the CI job in Jenkins
                                                                                          |
                                                                                          |
                                                                             +------------+-----------+
                                                                             |                        |
                                                                             | Secret Manifests Repo  |
                                                                             | (projection manifests) |
                                                                             |                        |
                                                                             +------------------------+

```

## Features

* Allow peer review and audit of exactly what namespaces have access to what secrets
* Allow developers to make requests for credentials access without exposing/accessing the credentials in question
* Enable applications to consume secrets in a structured format
* Allow encryption of secrets at in transit (and at rest) to the Kubernetes API
* Extract specific secrets from larger structured sources (YAML, JSON) via `JSONPath` notation
* Consume structured secrets in alternate formats at runtime (YAML/JSON/text), independent of source format
* Structured field extraction via `jsonpath` notation

## Examples

See docs at [docs/](/docs/examples.md)!

# Build

Builds are performed by Travis and Docker Hub. If you want to build this yourself, see below.

## Linux Native

```bash
$ GOOS=linux GOARCH=amd64 make
$ file bin/k8s-secret-projector
bin/k8s-forward-zone-generator: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, not stripped
```

## Docker

Docker images are built and pushed to the Hub automatically. ![Docker Build Status](https://img.shields.io/docker/build/tumblr/k8s-secret-projector.svg)

If you want to build this yourself, locally:

```bash
$ make docker
...
Successfully built 8ce7a5f5542c
Successfully tagged tumblr/k8s-secret-projector:v0.1.0-13-g1f564dc
```

## Plugins

We build plugins for use with the encryption module system. `make plugins` should do the needful.

The shared objects will be dropped in `/*.so`. They should be referenced by the ProjectionManifest's `Encryption.PluginFile` field.

# Development

See docs at [docs/hacking.md](/docs/hacking.md)!

# Maintainers

See [MAINTAINERS.md](/MAINTAINERS.md)

# License

[Apache 2.0](/LICENSE.txt)

Copyright 2018, Tumblr, Inc.
