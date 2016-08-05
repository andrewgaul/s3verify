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
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// newRemoveBucketReq - Fill in the dynamic fields of a DELETE request here.
func newRemoveBucketReq(config ServerConfig, bucketName string) (Request, error) {
	// removeBucketReq is a new DELETE bucket request.
	var removeBucketReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName.
	removeBucketReq.bucketName = bucketName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because DELETE requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	removeBucketReq.customHeader.Set("User-Agent", appUserAgent)
	removeBucketReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return removeBucketReq, nil
}

// removeBucketVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func removeBucketVerify(res *http.Response, expectedStatusCode int, errorResponse ErrorResponse) error {
	if err := verifyHeaderRemoveBucket(res.Header); err != nil {
		return err
	}
	if err := verifyStatusRemoveBucket(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyRemoveBucket(res.Body, errorResponse); err != nil {
		return err
	}
	return nil
}

// TODO: right now only checks for correctly deleted buckets...need to add in checks for 'failed' tests.

// verifyHeaderRemoveBucket - Check that the responses headers match the expected headers for a given DELETE Bucket request.
func verifyHeaderRemoveBucket(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyRemoveBucket - Check that the body of the response matches the expected body for a given DELETE Bucket request.
func verifyBodyRemoveBucket(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected.
		errResponse := ErrorResponse{}
		err := xmlDecoder(resBody, &errResponse)
		if err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error: %v", errResponse.Message)
			return err
		}
	}
	return nil
}

// verifyStatusRemoveBucket - Check that the status of the response matches the expected status for a given DELETE Bucket request.
func verifyStatusRemoveBucket(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode { // Successful DELETE request will result in 204 No Content.
		err := fmt.Errorf("Unexpected Status: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

//
func testRemoveBucketExists(config ServerConfig, curTest int, testBuckets []BucketInfo) bool {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket Exists):", curTest, globalTotalNumTest)
	for _, bucket := range testBuckets {
		// Spin the scanBar
		scanBar(message)
		// Generate the new DELETE bucket request.
		req, err := newRemoveBucketReq(config, bucket.Name)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin the scanBar
		scanBar(message)
		// Perform the request.
		res, err := config.execRequest("DELETE", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Spin the scanBar
		scanBar(message)
		if err := removeBucketVerify(res, 204, ErrorResponse{}); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin the scanBar
		scanBar(message)
	}
	printMessage(message, nil)
	return true
}

// mainRemoveBucketExistsUnPrepared - entry point for the RemoveBucket API test when --prepare was used and when the bucket exists.
func mainRemoveBucketExistsUnPrepared(config ServerConfig, curTest int) bool {
	// Remove all buckets.
	return testRemoveBucketExists(config, curTest, unpreparedBuckets)
}

//
func mainRemoveBucketExistsPrepared(config ServerConfig, curTest int) bool {
	// Only remove the buckets created by s3verify tests themselves. Not the buckets made by --prepare.
	return testRemoveBucketExists(config, curTest, s3verifyBuckets)
}

// Test the RemoveBucket API when the bucket does not exist.
func mainRemoveBucketDNE(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket DNE):", curTest, globalTotalNumTest)
	// Generate a random bucketName.
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	// Hardcode the expected error response.
	errResponse := ErrorResponse{
		Code:       "NoSuchBucket",
		Message:    "The specified bucket does not exist",
		BucketName: bucketName,
		Key:        "",
	}
	// Spin scanBar
	scanBar(message)
	// Generate a new DELETE bucket request for a bucket that does not exist.
	req, err := newRemoveBucketReq(config, bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := config.execRequest("DELETE", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	if err := removeBucketVerify(res, http.StatusNotFound, errResponse); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
