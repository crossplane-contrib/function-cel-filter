# function-cel-filter
[![CI](https://github.com/negz/function-cel-filter/actions/workflows/ci.yml/badge.svg)](https://github.com/negz/function-cel-filter/actions/workflows/ci.yml)

A [composition function][functions] that [filters][filter] matching composed
resources using [CEL expressions][cel].

Each filter:

* Matches composed resources by name using a regular expression.
* Specifies whether resources should be included using a CEL expression.

If a filter's CEL expression evaluates to true, Crossplane creates the matching
composed resources.

Filters only apply to matching composed resources. The function doesn't filter
composed resources that don't match a filter. 

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: function-template-go
spec:
  compositeTypeRef:
    apiVersion: example.crossplane.io/v1
    kind: NoSQL
  mode: Pipeline
  pipeline:
  - step: patch-and-transform
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:
      - name: table
        base:
          apiVersion: dynamodb.aws.upbound.io/v1beta1
          kind: Table
          metadata:
            name: crossplane-quickstart-database
          spec:
            forProvider:
              region: "us-east-2"
              writeCapacity: 1
              readCapacity: 1
              attribute:
                - name: S3ID
                  type: S
              hashKey: S3ID
      - name: bucket
        base:
          apiVersion: s3.aws.upbound.io/v1beta1
          kind: Bucket
          spec:
            forProvider:
              region: us-east-2
  - step: filter-composed-resources
    functionRef:
      name: function-cel-filter
    input:
      apiVersion: cel.fn.crossplane.io/v1beta1
      kind: Filters
      filters:
      # Only create the bucket if the XR's spec.export field is set to "S3".
      - name: bucket
        expression: observed.composite.resource.spec.export == "S3"
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[cel]: https://github.com/google/cel-spec
[filter]: https://en.wikipedia.org/wiki/Filter_(higher-order_function)
