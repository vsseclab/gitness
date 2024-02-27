// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openapi

import (
	"net/http"

	"github.com/harness/gitness/app/api/controller/user"
	"github.com/harness/gitness/app/api/request"
	"github.com/harness/gitness/app/api/usererror"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/gotidy/ptr"
	"github.com/swaggest/openapi-go/openapi3"
)

type createTokenRequest struct {
	user.CreateTokenInput
}

var queryParameterMembershipSpaces = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamQuery,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The substring by which the spaces the users is a member of are filtered."),
		Required:    ptr.Bool(false),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptrSchemaType(openapi3.SchemaTypeString),
			},
		},
	},
}

var queryParameterSortMembershipSpaces = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamSort,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The field by which the spaces the user is a member of are sorted."),
		Required:    ptr.Bool(false),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type:    ptrSchemaType(openapi3.SchemaTypeString),
				Default: ptrptr(enum.MembershipSpaceSortIdentifier),
				Enum:    enum.MembershipSpaceSort("").Enum(),
			},
		},
	},
}

// helper function that constructs the openapi specification
// for user account resources.
func buildUser(reflector *openapi3.Reflector) {
	opFind := openapi3.Operation{}
	opFind.WithTags("user")
	opFind.WithMapOfAnything(map[string]interface{}{"operationId": "getUser"})
	_ = reflector.SetRequest(&opFind, nil, http.MethodGet)
	_ = reflector.SetJSONResponse(&opFind, new(types.User), http.StatusOK)
	_ = reflector.SetJSONResponse(&opFind, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.Spec.AddOperation(http.MethodGet, "/user", opFind)

	opUpdate := openapi3.Operation{}
	opUpdate.WithTags("user")
	opUpdate.WithMapOfAnything(map[string]interface{}{"operationId": "updateUser"})
	_ = reflector.SetRequest(&opUpdate, new(user.UpdateInput), http.MethodPatch)
	_ = reflector.SetJSONResponse(&opUpdate, new(types.User), http.StatusOK)
	_ = reflector.SetJSONResponse(&opUpdate, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.Spec.AddOperation(http.MethodPatch, "/user", opUpdate)

	opToken := openapi3.Operation{}
	opToken.WithTags("user")
	opToken.WithMapOfAnything(map[string]interface{}{"operationId": "createToken"})
	_ = reflector.SetRequest(&opToken, new(createTokenRequest), http.MethodPost)
	_ = reflector.SetJSONResponse(&opToken, new(types.TokenResponse), http.StatusCreated)
	_ = reflector.SetJSONResponse(&opToken, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.Spec.AddOperation(http.MethodPost, "/user/token", opToken)

	opMemberSpaces := openapi3.Operation{}
	opMemberSpaces.WithTags("user")
	opMemberSpaces.WithMapOfAnything(map[string]interface{}{"operationId": "membershipSpaces"})
	opMemberSpaces.WithParameters(
		queryParameterMembershipSpaces,
		queryParameterOrder, queryParameterSortMembershipSpaces,
		queryParameterPage, queryParameterLimit)
	_ = reflector.SetRequest(&opMemberSpaces, struct{}{}, http.MethodGet)
	_ = reflector.SetJSONResponse(&opMemberSpaces, new([]types.MembershipSpace), http.StatusOK)
	_ = reflector.SetJSONResponse(&opMemberSpaces, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.Spec.AddOperation(http.MethodGet, "/user/memberships", opMemberSpaces)
}
