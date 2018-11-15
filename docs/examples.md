# Examples

Using example secret projection mappings in `examples/manifests/`, like `myservice-example.yaml`, we can
extract some secrets from structure JSON source files, and create a Kubernetes
Secret that contains those secrets:

```yaml
---
name: someservice
namespace: myservice
repo: production
data:
- name: api-key
  source:
    json: applications/someservice/creds.json
    jsonpath: $.apikey
- name: database-master
  source:
    json: mysql/platform/creds.json
    jsonpath: $.someservice.creds.host
- name: database-username
  source:
    json: mysql/platform/creds.json
    jsonpath: $.someservice.creds.username
- name: database-password
  source:
    json: mysql/platform/creds.json
    jsonpath: $.someservice.creds.password
- name: a-file.txt
  source:
    raw: applications/someservice/a-file.txt
- name: secret_password
  source:
    yaml: applications/someservice/secrets.yaml
    jsonpath: $.secrets.password
- name: secret_keys.json
  source:
    yaml: applications/someservice/secrets.yaml
    format: json
    jsonpaths: 
      username: $.secrets.username    
      password: $.secrets.password    
- name: secret_keys.yaml
  source:
    yaml: applications/someservice/secrets.yaml
    format: yaml
    jsonpaths: 
      username: $.secrets.username    
      password: $.secrets.password    
```

Then, we apply the tool on these projection mapping manifests into Kubernetes secrets in `tmp/` like:

```bash
$ make
▶ running gofmt…
▶ running golint…
▶ building executable…
$ ./bin/k8s-secret-projector -creds-repo=development=example/creds/,staging=example/creds/,production=example/creds/ -manifests example/manifests/ -debug-show-secrets -output tmp/
2017/11/10 12:13:59 Loading from config: [./bin/k8s-secret-projector -creds-repo=development=example/creds/,staging=example/creds/,production=example/creds/ -manifests example/manifests/ -debug-show-secrets -output tmp/]
2017/11/10 12:13:59 creds development path: example/creds/
2017/11/10 12:13:59 creds staging path: example/creds/
2017/11/10 12:13:59 creds production path: example/creds/
2017/11/10 12:13:59 manifests: example/manifests/
2017/11/10 12:13:59 Loaded manifest: json-tests/test1:production{single-json-key:json:object1.json,another-json-key:json:object1.json,an-int:json:object1.json}
2017/11/10 12:13:59 Loaded manifest: raw-test1/test2:production{raw-file:raw:raw1.txt}
2017/11/10 12:13:59 Loaded 2 manifests
2017/11/10 12:13:59 writing json-tests/test1 secret to tmp/1510334039-active_json-tests-test1.yaml...
2017/11/10 12:13:59 writing raw-test1/test2 secret to tmp/1510334039-active_raw-test1-test2.yaml...
$ for f in $(ls tmp) ; do echo $f ; cat tmp/$f ; done
1510178393-json-tests-test1.yaml
apiVersion: v1
data:
  another-json-key: Zm9v
  single-json-key: cGFTc3cwcmQh
kind: Secret
metadata:
  creationTimestamp: null
  name: test1
  namespace: json-tests
type: Opaque
1510178393-active_raw-test1-test2.yaml
apiVersion: v1
data:
  raw-file: aGVsbG8KdGhpcyBpcyBhIHJhdyBmaWxlCg==
kind: Secret
metadata:
  creationTimestamp: null
  name: test2
  namespace: raw-test1
type: Opaque
```

Alternatively, you can run the same thing with docker:
```bash
$ [[ -d ./output ]] || mkdir output ; docker run -it --rm \
  -v $(pwd):/manifests \
  -v $(pwd)/example/creds/:/credentials/production \
  -v $(pwd)/example/creds/:/credentials/development \
  -v $(pwd)/example/creds/:/credentials/staging \
  -v $(pwd)/output:/output \
  tumblr/k8s-secret-projector:latest \
    -creds-repo=staging=/credentials/staging,production=/credentials/production,development=/credentials/development \
    -manifests /manifests \
    -debug-show-secrets \
    -output /output
```
And check `./output` for the generated secrets.

## Good to know

The `k8s-secret-projector` makes the assumption that you keep multiple credentials repos. For this example, we are going to use three repositories, keyed by an identifier:

* `production`: `--creds-repo=production=/credentials/production`
* `staging`: `--creds-repo=staging=/credentials/staging`
* `development`: `--creds-repo=development=/credentials/development`

You may name your creds repos as you choose; please note that the identifiers must match with a projection manifest's `repo:` field. The repo fields used must correspond to the flag `--creds-repo` parameters, which are `identifier=directory[,identifier=directory...]`. The `repo:` field of a projection manifest tells the projector which repository to source its credentials; each source path is relative to the specific repo directory passed at runtime.

If you only have a single monolithic creds repo, you can use `--creds-repo=production=/path/to/repo`. Just make sure all of your projection manifests use the proper `repo: production` setting :)


# More examples!

## Simple Raw File Projection

Do you have files in your `production` repo you need access to in your Deployment? This example uses a raw file to create a Secret. Note, the repo passed to `--creds-repo=production=/some/git/repo`. This will project a single file called `s3.key` in Kubernetes Secret named `aws-s3-key`:

```yaml
name: aws-s3-key
namespace: myteam
repo: production
data:
- name: s3.key
  source:
    raw: applications/aws/s3key.txt
```

## Simple JSON Field Extraction

Do you have JSON structured file in the `production` creds repo, and want a password out of it? We can extract the value from structured data and make a secret containing just the value! This assumes the `credentials.json` looks like: `{"s3":{"key":"passW0rD!"},"aws":{}}`. (https://github.com/oliveagle/jsonpath for notes on jsonpath syntax)

```yaml
name: aws-s3-key
namespace: myteam
repo: production
data:
- name: s3.key
  source:
    json: applications/aws/credentials.json
    jsonpath: $.s3.key
```

## Simple YAML Field Extraction

We can extract fields from YAML secret sources just like JSON! Just specify a `yaml:` source instead of `json:`, and use the same `jsonpath` notation (https://github.com/oliveagle/jsonpath)!

```yaml
name: aws-s3-key
namespace: myteam
repo: production
data:
- name: s3.key
  source:
    yaml: applications/aws/credentials.yaml
    jsonpath: $.s3.key
```

## Structured Subset Projection

Ok, so you have some structured source data (json, yaml, whatever) that you want to extract multiple fields from, and project into a structured format. We can do that, too! Assume a `credentials.json` that looks like the following (in creds repo `production`):

```json
{
  "aws": {
    "key": "somethignSekri7T!",
    "region": "US-West-2"
  },
  "s3": {
    "key": "passW0rD!",
    "region": "US-East-1"
  },
  "redshift": {
    "key": "we dont want to project redshift key into our app",
    "database": "whatever",
  }
}
```
We can extract a few fields, and build a new output yaml file for our application to consume! Note the `format: yaml` that tells the projector to output yaml, instead of the assumed JSON (assumed the output is the same as input format).

```yaml
name: aws-credentials
namespace: myteam
repo: production
data:
- name: amazon.yaml
  source:
    format: yaml
    json: applications/aws/credentials.json
    jsonpaths:
      s3key: $.s3.key
      awskey: $.aws.key
      awsregion: $.aws.region
```

This secret projection manifest would result in a _dynamically generated_ `YAML` file being generated at `amazon.yaml` that looks like:

```yaml
---
s3key: "passW0rD!"
awskey: "somethignSekri7T!"
awsregion: "US-West-2"
```

NOTE: Supported `format:` keys are `yaml` or `json`, and are only valid when you use the `jsonpaths` key to specify multiple fields to extract from a source. If you extract multiple fields from a `json` source, it will default to projecting them as a JSON object (and yaml for yaml), unless you override with the `format` key.

## Encryption

Encrypting each individual data item is possible.

1. Select the encryption module (here its the "plugin" module) with `$.encryption.module`
2. Select the desired plugin for encryption (`cbc.so`, but you can make your own modules)
3. Tell the encryption module whether to also include the decryption_keys with the `Secret`: `$.encryption.include_decryption_keys`
4. Specify `encrypt: true` for each item you want to be encrypted.

NOTE: `include_decryption_keys` will create a `keys_${n}.json` item for the keys that the encryption module decides are relevant to decrypt your application. The number of `keys_*.json` files injected is controlled by the encryption module in use. See [plugins/encryption/cbc](/plugins/encryption/cbc/main.go) for an example plugin.

```yaml
name: test-json-subset
namespace: json-tests
repo: production
encryption:
  module: plugin
  plugin-path: cbc.so
  include_decryption_keys: true
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

