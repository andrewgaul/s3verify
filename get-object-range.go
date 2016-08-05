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
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// newGetObjectRangeReq - Create a new GET object range request.
func newGetObjectRangeReq(config ServerConfig, bucketName, objectName string, startRange, endRange int64) (Request, error) {
	// getObjectRangeReq - a new HTTP request for a GET object with a specific range request.
	var getObjectRangeReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	getObjectRangeReq.bucketName = bucketName
	getObjectRangeReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getObjectRangeReq.customHeader.Set("Range", "bytes="+strconv.FormatInt(startRange, 10)+"-"+strconv.FormatInt(endRange, 10))
	getObjectRangeReq.customHeader.Set("User-Agent", appUserAgent)
	getObjectRangeReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	return getObjectRangeReq, nil
}

// Test a GET object request with a range header set.
func testGetObjectRange(config ServerConfig, curTest int, bucketName string, testObjects []*ObjectInfo) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (Range):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	rand.Seed(time.Now().UnixNano())
	for _, object := range testObjects {
		// Spin scanBar
		scanBar(message)
		startRange := rand.Int63n(object.Size)
		endRange := rand.Int63n(int64(object.Size-startRange)) + startRange
		// Create new GET object range request...testing range.
		req, err := newGetObjectRangeReq(config, bucketName, object.Key, startRange, endRange)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("GET", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		bufRange := object.Body[startRange : endRange+1]
		// Verify the response...these checks do not check the headers yet.
		if err := getObjectVerify(res, bufRange, http.StatusPartialContent); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)

	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// mainGetObjectRangePrepared - entry point for the GetObject with range header set and --prepare used.
func mainGetObjectRangePrepared(config ServerConfig, curTest int) bool {
	bucketName := s3verifyBuckets[0].Name
	return testGetObjectRange(config, curTest, bucketName, s3verifyObjects)
}

//
func mainGetObjectRangeUnPrepared(config ServerConfig, curTest int) bool {
	bucketName := unpreparedBuckets[0].Name
	return testGetObjectRange(config, curTest, bucketName, objects)
}
