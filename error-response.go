/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/xml"
	"net/http"
)

/* **** SAMPLE ERROR RESPONSE ****
<?xml version="1.0" encoding="UTF-8"?>
<Error>
   <Code>AccessDenied</Code>
   <Message>Access Denied</Message>
   <BucketName>bucketName</BucketName>
   <Key>objectName</Key>
   <RequestId>F19772218238A85A</RequestId>
   <HostId>GuWkjyviSiGHizehqpmsD1ndz5NClSP19DOT+s2mv7gXGQ8/X1lhbDGiIJEXpGFD</HostId>
</Error>
*/

// ErrorResponse - Is the typed error returned by all API operations.
type ErrorResponse struct {
	XMLName    xml.Name `xml:"Error" json:"-"`
	Code       string
	Message    string
	BucketName string
	Key        string
	RequestID  string `xml:"RequestId"`
	HostID     string `xml:"HostId"`

	// Region where the bucket is located. This header is returned
	// only in HEAD bucket and ListObjects response.
	Region string
}

// ToErrorResponse - Returns parsed ErrorResponse struct from body and
// http headers.
//
// For example:
//
//   import s3 "github.com/minio/minio-go"
//   ...
//   ...
//   reader, stat, err := s3.GetObject(...)
//   if err != nil {
//      resp := s3.ToErrorResponse(err)
//   }
//   ...
func ToErrorResponse(err error) ErrorResponse {
	switch err := err.(type) {
	case ErrorResponse:
		return err
	default:
		return ErrorResponse{}
	}
}

// Error - Returns HTTP error string
func (e ErrorResponse) Error() string {
	return e.Message
}

// Common string for errors to report issue location in unexpected
// cases.
const (
	reportIssue = "Please report this issue at https://github.com/minio/s3verify/issues."
)

// httpRespToErrorResponse returns a new encoded ErrorResponse
// structure as error.
func httpRespToErrorResponse(resp *http.Response, bucketName, objectName string) error {
	if resp == nil {
		msg := "Response is empty. " + reportIssue
		return ErrInvalidArgument(msg)
	}
	var errResp ErrorResponse
	err := xmlDecoder(resp.Body, &errResp)
	// Xml decoding failed with no body, fall back to HTTP headers.
	if err != nil {
		switch resp.StatusCode {
		case http.StatusNotFound:
			if objectName == "" {
				errResp = ErrorResponse{
					Code:       "NoSuchBucket",
					Message:    "The specified bucket does not exist.",
					BucketName: bucketName,
					RequestID:  resp.Header.Get("x-amz-request-id"),
					HostID:     resp.Header.Get("x-amz-id-2"),
					Region:     resp.Header.Get("x-amz-bucket-region"),
				}
			} else {
				errResp = ErrorResponse{
					Code:       "NoSuchKey",
					Message:    "The specified key does not exist.",
					BucketName: bucketName,
					Key:        objectName,
					RequestID:  resp.Header.Get("x-amz-request-id"),
					HostID:     resp.Header.Get("x-amz-id-2"),
					Region:     resp.Header.Get("x-amz-bucket-region"),
				}
			}
		case http.StatusForbidden:
			errResp = ErrorResponse{
				Code:       "AccessDenied",
				Message:    "Access Denied.",
				BucketName: bucketName,
				Key:        objectName,
				RequestID:  resp.Header.Get("x-amz-request-id"),
				HostID:     resp.Header.Get("x-amz-id-2"),
				Region:     resp.Header.Get("x-amz-bucket-region"),
			}
		case http.StatusConflict:
			errResp = ErrorResponse{
				Code:       "Conflict",
				Message:    "Bucket not empty.",
				BucketName: bucketName,
				RequestID:  resp.Header.Get("x-amz-request-id"),
				HostID:     resp.Header.Get("x-amz-id-2"),
				Region:     resp.Header.Get("x-amz-bucket-region"),
			}
		default:
			errResp = ErrorResponse{
				Code:       resp.Status,
				Message:    resp.Status,
				BucketName: bucketName,
				RequestID:  resp.Header.Get("x-amz-request-id"),
				HostID:     resp.Header.Get("x-amz-id-2"),
				Region:     resp.Header.Get("x-amz-bucket-region"),
			}
		}
	}
	return errResp
}

// ErrInvalidArgument - Invalid argument response.
func ErrInvalidArgument(message string) error {
	return ErrorResponse{
		Code:      "InvalidArgument",
		Message:   message,
		RequestID: "minio",
	}
}

// ErrInvalidBucketName - Invalid bucket name response.
func ErrInvalidBucketName(message string) error {
	return ErrorResponse{
		Code:      "InvalidBucketName",
		Message:   message,
		RequestID: "minio",
	}
}

// ErrInvalidObjectName - Invalid object name response.
func ErrInvalidObjectName(message string) error {
	return ErrorResponse{
		Code:      "NoSuchKey",
		Message:   message,
		RequestID: "minio",
	}
}