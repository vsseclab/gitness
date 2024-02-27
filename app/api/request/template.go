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

package request

import (
	"net/http"
	"net/url"
)

const (
	PathParamTemplateRef  = "template_ref"
	PathParamTemplateType = "template_type"
)

func GetTemplateRefFromPath(r *http.Request) (string, error) {
	rawRef, err := PathParamOrError(r, PathParamTemplateRef)
	if err != nil {
		return "", err
	}

	// paths are unescaped
	return url.PathUnescape(rawRef)
}

func GetTemplateTypeFromPath(r *http.Request) (string, error) {
	templateType, err := PathParamOrError(r, PathParamTemplateType)
	if err != nil {
		return "", err
	}

	return templateType, nil
}
