# Encryption

The secrets projector can encrypt your `Secret`s before they hit the k8s API. This is useful if you are running your apps in an environment where you dont trust your Kubernetes provider, and want to ensure your secrets are encrypted at rest/in flight to/from etcd. This requires some extra legwork though, because your application must know how to decrypt these secrets before use.

## Example

### CBC

We ship a reference CBC encryption module with the projector. This currently has support for `aes` cipher, but more may be added in the future.

You can configure encryption of your secrets before the secret is created in the Kubernetes API. NOTE: this requires that your application knows how to decrypt your secrets before use :)

The following will encrypt the projected secrets with CBC/AES with MD5 hashing of the password. Note, this requires the `k8s-secret-projector` to be launched with the encryption and decryption keys. (`--creds-encryption-key` and `--creds-key-decryption-key`)


```yaml
---
name: test-json-subset
namespace: json-tests
repo: internal
encryption:
  module: cbc
  include_decryption_keys: true
  params:
    cipher: aes
    hash: md5
data:
- name: secrets.json.enc
  encrypt: true
  source:
    json: object1.json
    jsonpaths:
      key1: $.nesting.key1
      float: $.nesting.float
      list: $.nesting.list
```

# Encryption and Decryption Keys

TODO: clarify how `--creds-encryption-key` and `--creds-key-decryption-key` works, format for keys, etc
