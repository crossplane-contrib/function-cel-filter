package main

import (
	"context"
	"regexp"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/crossplane-contrib/function-cel-filter/input/v1beta1"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
	env *cel.Env
}

// NewFunction creates a new Function with a CEL environment.
func NewFunction(log logging.Logger) (*Function, error) {
	env, err := cel.NewEnv(
		cel.Types(&fnv1beta1.State{}, &structpb.Struct{}),
		cel.Variable("observed", cel.ObjectType("apiextensions.fn.proto.v1beta1.State")),
		cel.Variable("desired", cel.ObjectType("apiextensions.fn.proto.v1beta1.State")),
		cel.Variable("context", cel.ObjectType("google.protobuf.Struct")),

		ext.Encoders(),
		ext.Lists(),
		ext.Math(),
		cel.ExtendedValidations(),
		cel.EagerlyValidateDeclarations(true),
		cel.DefaultUTCTimeZone(true),
		cel.CrossTypeNumericComparisons(true),
		cel.OptionalTypes(),
		ext.Strings(ext.StringsVersion(2)),
		ext.Sets(),
	)
	return &Function{log: log, env: env}, errors.Wrap(err, "cannot create CEL environment")
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Filters{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get filters from %T", req))
		return rsp, nil
	}

	regexps := make([]*regexp.Regexp, len(in.Filters))
	celexps := make([]bool, len(in.Filters))

	for i := range in.Filters {
		// Compile this filter's regular expression.
		expr := "^" + in.Filters[i].Name + "$"
		re, err := regexp.Compile(expr)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot compile regular expression %q for filter %d", expr, i))
			return rsp, nil
		}
		regexps[i] = re

		// Evaluate this filter's CEL expression.
		expr = in.Filters[i].Expression
		include, err := Evaluate(f.env, req, in.Filters[i].Expression)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot evaluate CEL expression %q for filter %d", expr, i))
			return rsp, nil
		}
		celexps[i] = include
	}

	for name := range rsp.GetDesired().GetResources() {
		log := f.log.WithValues("resource-name", name)
		for i := range in.Filters {
			log = log.WithValues(
				"filter-index", i,
				"filter-name", in.Filters[i].Name,
				"filter-expression", in.Filters[i].Expression,
			)
			if !regexps[i].MatchString(name) {
				log.Debug("Not filtering desired composed resource: composed resource name does not match filter name")
				continue
			}

			if include := celexps[i]; include {
				log.Debug("Not filtering desired composed resource: CEL expression evaluated to true")
				continue
			}

			log.Info("Filtering desired composed resource: CEL expression evaluated to false")
			delete(rsp.GetDesired().GetResources(), name)
		}
	}

	return rsp, nil
}

// Evaluate the supplied CEL expression.
func Evaluate(env *cel.Env, req *fnv1beta1.RunFunctionRequest, expression string) (bool, error) {
	ast, iss := env.Parse(expression)
	if iss.Err() != nil {
		return false, errors.Wrap(iss.Err(), "cannot parse expression")
	}

	// Type-check the expression for correctness.
	checked, iss := env.Check(ast)
	if iss.Err() != nil {
		return false, errors.Wrap(iss.Err(), "cannot type-check expression")
	}

	if !checked.OutputType().IsExactType(cel.BoolType) {
		return false, errors.Errorf("expression %q must return a boolean, but will return %s instead", expression, checked.OutputType())
	}

	program, err := env.Program(checked)
	if err != nil {
		return false, errors.Wrap(err, "cannot create CEL program")
	}

	result, _, err := program.Eval(map[string]any{
		"observed": req.GetObserved(),
		"desired":  req.GetDesired(),
		"context":  req.GetContext(),
	})
	if err != nil {
		return false, errors.Wrap(err, "cannot evaluate CEL program")
	}

	ret, ok := result.Value().(bool)
	if !ok {
		return false, errors.Wrap(err, "expression did not return a bool")
	}

	return ret, nil
}
