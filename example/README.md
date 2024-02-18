# Example manifests

You can run your function locally and test it using `crossplane beta render`
with these example manifests.

```shell
# Run the function locally
$ go run . --insecure --debug
```

```shell
# Then, in another terminal, call it with these example manifests
$ crossplane beta render xr.yaml composition.yaml functions.yaml -r
---
apiVersion: example.crossplane.io/v1
kind: NoSQL
metadata:
  name: example-xr
---
apiVersion: dynamodb.aws.upbound.io/v1beta1
kind: Table
metadata:
  annotations:
    crossplane.io/composition-resource-name: table
  generateName: example-xr-
  labels:
    crossplane.io/composite: example-xr
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: NoSQL
    name: example-xr
    uid: ""
spec:
  forProvider:
    attribute:
    - name: S3ID
      type: S
    hashKey: S3ID
    readCapacity: 1
    region: us-east-2
    writeCapacity: 1
```
