# function-cel-filter
[![CI](https://github.com/negz/function-cel-filter/actions/workflows/ci.yml/badge.svg)](https://github.com/negz/function-cel-filter/actions/workflows/ci.yml)

A [composition function][functions] that can filter desired composed resources
produced by previous functions in the pipeline using CEL expressions.

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: function-template-go
spec:
  compositeTypeRef:
    apiVersion: example.crossplane.io/v1
    kind: XR
  mode: Pipeline
  pipeline:
  # TODO(negz): Replace me with function-dummy.
  - step: produce-composed-resources
    functionRef:
      name: some-composition-function
  - step: filter-composed-resources
    functionRef:
      name: function-cel-filter
    input:
      apiVersion: cel.fn.crossplane.io/v1beta1
      kind: Filters
      filters:
      # Remove the composed resource named a-desired-composed-resource
      # from the function pipeline if the XR has spec.widgets == 42.
      - name: a-desired-composed-resource
        expression: observed.composed.resource.spec.widgets == 42
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[go]: https://go.dev
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[docker]: https://www.docker.com
[cli]: https://docs.crossplane.io/latest/cli