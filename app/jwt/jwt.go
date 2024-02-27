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

package jwt

import (
	"time"

	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

const (
	issuer = "Gitness"
)

// Claims defines gitness jwt claims.
type Claims struct {
	jwt.StandardClaims

	PrincipalID int64 `json:"pid,omitempty"`

	Token      *SubClaimsToken      `json:"tkn,omitempty"`
	Membership *SubClaimsMembership `json:"ms,omitempty"`
}

// SubClaimsToken contains information about the token the JWT was created for.
type SubClaimsToken struct {
	Type enum.TokenType `json:"typ,omitempty"`
	ID   int64          `json:"id,omitempty"`
}

// SubClaimsMembership contains the ephemeral membership the JWT was created with.
type SubClaimsMembership struct {
	Role    enum.MembershipRole `json:"role,omitempty"`
	SpaceID int64               `json:"sid,omitempty"`
}

// GenerateForToken generates a jwt for a given token.
func GenerateForToken(token *types.Token, secret string) (string, error) {
	var expiresAt int64
	if token.ExpiresAt != nil {
		expiresAt = *token.ExpiresAt
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec not millisec
			IssuedAt:  token.IssuedAt / 1000,
			ExpiresAt: expiresAt / 1000,
		},
		PrincipalID: token.PrincipalID,
		Token: &SubClaimsToken{
			Type: token.Type,
			ID:   token.ID,
		},
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", errors.Wrap(err, "Failed to sign token")
	}

	return res, nil
}

// GenerateWithMembership generates a jwt with the given ephemeral membership.
func GenerateWithMembership(
	principalID int64,
	spaceID int64,
	role enum.MembershipRole,
	lifetime time.Duration,
	secret string,
) (string, error) {
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(lifetime)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer: issuer,
			// times required to be in sec
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: expiresAt.Unix(),
		},
		PrincipalID: principalID,
		Membership: &SubClaimsMembership{
			SpaceID: spaceID,
			Role:    role,
		},
	})

	res, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return "", errors.Wrap(err, "Failed to sign token")
	}

	return res, nil
}
