package main

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"EmptyFiltersDoesNothing": {
			reason: "The function should return all desired resources if there are no filters",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					Input: resource.MustStructJSON(`{
						"apiVersion": "filters.cel.crossplane.io/v1beta1",
						"kind": "Filters"
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
				},
			},
		},
		"BasicFilter": {
			reason: "If the filter name matches a resource, it should only be included if the CEL expression evaluates to true",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					// The first filter matches the resources but it evaluates
					// to true, so it won't filter the resources. However the
					// second filter will, because it also matches the resources
					// and evaluates to false.
					Input: resource.MustStructJSON(`{
						"apiVersion": "filters.cel.crossplane.io/v1beta1",
						"kind": "Filters",
						"filters": [
							{
								"name": "matching-.*",
								"expression": "observed.composite.resource.spec.watchers == 42"
							},
							{
								"name": "matching-.*",
								"expression": "observed.composite.resource.spec.widgets == 88"
							}
						]
					}`),
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"spec": {
									"watchers": 42,
									"widgets": 42
								}
							}`),
						},
					},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"matching-resource-a":   {},
							"matching-resource-b":   {},
							"non-matching-resource": {},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							// matching-resource-a was filtered.
							// matching-resource-b was filtered.
							"non-matching-resource": {},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f, _ := NewFunction(logging.NewNopLogger())
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
