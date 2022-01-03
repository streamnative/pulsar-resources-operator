// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package admin

import (
	pulsarcli "github.com/streamnative/pulsarctl/pkg/cli"
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
	cliError, ok := err.(pulsarcli.Error)
	if !ok {
		// can't determine error reason as can't convert to a cli error
		return ReasonUnknown
	}
	return Reason(cliError.Code)
}

// IsNotFound returns true if the error indicates the resource is not found on server
func IsNotFound(err error) bool {
	return ErrorReason(err) == ReasonNotFound
}

// IsAlreadyExist returns true if the error indicates the resource already exist
func IsAlreadyExist(err error) bool {
	return ErrorReason(err) == ReasonAlreadyExist
}

// IsInternalServerError returns true if the error indicates the resource already exist
func IsInternalServerError(err error) bool {
	return ErrorReason(err) == ReasonInternalServerError
}
