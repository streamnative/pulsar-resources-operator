// Copyright 2025 StreamNative
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package admin

import (
	"errors"
	"net"
	"strings"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/rest"
)

// Reason indicates the status code
type Reason int

const (
	// ReasonUnauthorized means need to authenticate to perform the operation
	ReasonUnauthorized Reason = 401

	// ReasonForbidden means don't have admin permission for the operation
	ReasonForbidden Reason = 403

	// ReasonNotFound means a resource is not found in Pulsar
	ReasonNotFound Reason = 404

	// ReasonAlreadyExist means a resource already exist in Pulsar
	ReasonAlreadyExist Reason = 409

	// ReasonInvalidParameter means a resource already exist in Pulsar
	// Status code 412
	ReasonInvalidParameter Reason = 412

	// ReasonInternalServerError means Pulsar server fail to handle the request
	// Status code 500
	ReasonInternalServerError Reason = 500

	// ReasonUnknown means error reason is not clear
	ReasonUnknown Reason = 0
)

// ErrorReason returns the HTTP status code for the error
func ErrorReason(err error) Reason {
	if err == nil {
		return ReasonUnknown
	}

	var restErrPtr *rest.Error
	if errors.As(err, &restErrPtr) && restErrPtr != nil {
		return Reason(restErrPtr.Code)
	}

	var restErr rest.Error
	if errors.As(err, &restErr) {
		return Reason(restErr.Code)
	}

	// can't determine error reason as can't convert to a cli error
	return ReasonUnknown
}

// IsNotFound returns true if the error indicates the resource is not found on server
func IsNotFound(err error) bool {
	return ErrorReason(err) == ReasonNotFound
}

// IsAlreadyExist returns true if the error indicates the resource already exist
func IsAlreadyExist(err error) bool {
	if err == nil {
		return false
	}

	reason := ErrorReason(err)
	if reason == ReasonAlreadyExist {
		return true
	}
	if reason == ReasonInvalidParameter {
		return isAlreadyExistsMessage(err)
	}
	if reason != ReasonUnknown {
		return false
	}
	return isAlreadyExistsMessage(err)
}

func isAlreadyExistsMessage(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "already exist")
}

// IsInternalServerError returns true if the error indicates the resource already exist
func IsInternalServerError(err error) bool {
	return ErrorReason(err) == ReasonInternalServerError
}

// IsNoSuchHostError returns true if operator cannot connect the resource host
func IsNoSuchHostError(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr) && dnsErr.Err == "no such host"
}
