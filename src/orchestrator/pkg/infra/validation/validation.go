/*
 * Copyright 2020 Intel Corporation, Inc
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

package validation

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/util/validation"
)

func IsTarGz(r io.Reader) error {
	//Check if it is a valid gz
	gzf, err := gzip.NewReader(r)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid gzip format")
	}

	//Check if it is a valid tar file
	//Unfortunately this can only be done by inspecting all the tar contents
	tarR := tar.NewReader(gzf)
	first := true

	for true {
		header, err := tarR.Next()

		if err == io.EOF {
			//Check if we have just a gzip file without a tar archive inside
			if first {
				return pkgerrors.New("Empty or non-existant Tar file found")
			}
			//End of archive
			break
		}

		if err != nil {
			return pkgerrors.Errorf("Error reading tar file %s", err.Error())
		}

		//Check if files are of type directory and regular file
		if header.Typeflag != tar.TypeDir &&
			header.Typeflag != tar.TypeReg {
			return pkgerrors.Errorf("Unknown header in tar %s, %s",
				header.Name, string(header.Typeflag))
		}

		first = false
	}

	return nil
}

func IsIpv4Cidr(cidr string) error {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil || ip.To4() == nil {
		return pkgerrors.Wrapf(err, "could not parse ipv4 cidr %v", cidr)
	}
	return nil
}

func IsIp(ip string) error {
	addr := net.ParseIP(ip)
	if addr == nil {
		return pkgerrors.Errorf("invalid ip address %v", ip)
	}
	return nil
}

func IsIpv4(ip string) error {
	addr := net.ParseIP(ip)
	if addr == nil || addr.To4() == nil {
		return pkgerrors.Errorf("invalid ipv4 address %v", ip)
	}
	return nil
}

func IsMac(mac string) error {
	_, err := net.ParseMAC(mac)
	if err != nil {
		return pkgerrors.Errorf("invalid MAC address %v", mac)
	}
	return nil
}

// default name check - matches valid label value with addtion that length > 0
func IsValidName(name string) []string {
	var errs []string

	errs = validation.IsValidLabelValue(name)
	if len(name) == 0 {
		errs = append(errs, "name must have non-zero length")
	}
	return errs
}

const VALID_NAME_STR string = "NAME"

var validNameRegEx = regexp.MustCompile("^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$")

const VALID_ALPHA_STR string = "ALPHA"

var validAlphaStrRegEx = regexp.MustCompile("^[A-Za-z]*$")

const VALID_ALPHANUM_STR string = "ALPHANUM"

var validAlphaNumStrRegEx = regexp.MustCompile("^[A-Za-z0-9]*$")

// doesn't verify valid base64 length - just checks for proper base64 characters
const VALID_BASE64_STR string = "BASE64"

var validBase64StrRegEx = regexp.MustCompile("^[A-Za-z0-9+/]+={0,2}$")

const VALID_ANY_STR string = "ANY"

var validAnyStrRegEx = regexp.MustCompile("(?s)^.*$")

// string check - validates for conformance to provided lengths and specified content
// min and max - the string
// if format string provided - check against matching predefined
func IsValidString(str string, min, max int, format string) []string {
	var errs []string

	if min > max {
		errs = append(errs, "Invalid string length constraints - min is greater than max")
		return errs
	}

	if len(str) < min {
		errs = append(errs, "string length is less than the minimum constraint")
		return errs
	}
	if len(str) > max {
		errs = append(errs, "string length is greater than the maximum constraint")
		return errs
	}

	switch format {
	case VALID_ALPHA_STR:
		if !validAlphaStrRegEx.MatchString(str) {
			errs = append(errs, "string does not match the alpha only constraint")
		}
	case VALID_ALPHANUM_STR:
		if !validAlphaNumStrRegEx.MatchString(str) {
			errs = append(errs, "string does not match the alphanumeric only constraint")
		}
	case VALID_NAME_STR:
		if !validNameRegEx.MatchString(str) {
			errs = append(errs, "string does not match the valid k8s name constraint")
		}
	case VALID_BASE64_STR:
		if !validBase64StrRegEx.MatchString(str) {
			errs = append(errs, "string does not match the valid base64 characters constraint")
		}
		if len(str)%4 != 0 {
			errs = append(errs, "base64 string length should be a multiple of 4")
		}
	case VALID_ANY_STR:
		if !validAnyStrRegEx.MatchString(str) {
			errs = append(errs, "string does not match the any characters constraint")
		}
	default:
		// invalid string format supplied
		errs = append(errs, "an invalid string constraint was supplied")
	}

	return errs
}

// validate that label conforms to kubernetes label conventions
//  general label format expected is:
//  "<labelprefix>/<labelname>=<Labelvalue>"
//  where labelprefix matches DNS1123Subdomain format
//        labelname matches DNS1123Label format
//
// Input labels are allowed to  match following formats:
//  "<DNS1123Subdomain>/<DNS1123Label>=<Labelvalue>"
//  "<DNS1123Label>=<LabelValue>"
//  "<LabelValue>"
func IsValidLabel(label string) []string {
	var labelerrs []string

	expectLabelName := false
	expectLabelPrefix := false

	// split label up into prefix, name and value
	// format:  prefix/name=value
	var labelprefix, labelname, labelvalue string

	kv := strings.SplitN(label, "=", 2)
	if len(kv) == 1 {
		labelprefix = ""
		labelname = ""
		labelvalue = kv[0]
	} else {
		pn := strings.SplitN(kv[0], "/", 2)
		if len(pn) == 1 {
			labelprefix = ""
			labelname = pn[0]
		} else {
			labelprefix = pn[0]
			labelname = pn[1]
			expectLabelPrefix = true
		}
		labelvalue = kv[1]
		// if "=" was in the label input, then expect a non-zero length name
		expectLabelName = true
	}

	// check label prefix validity - prefix is optional
	if len(labelprefix) > 0 {
		errs := validation.IsDNS1123Subdomain(labelprefix)
		if len(errs) > 0 {
			labelerrs = append(labelerrs, "Invalid label prefix - label=["+label+"%], labelprefix=["+labelprefix+"], errors: ")
			for _, err := range errs {
				labelerrs = append(labelerrs, err)
			}
		}
	} else if expectLabelPrefix {
		labelerrs = append(labelerrs, "Invalid label prefix - label=["+label+"%], labelprefix=["+labelprefix+"]")
	}
	if expectLabelName {
		errs := validation.IsDNS1123Label(labelname)
		if len(errs) > 0 {
			labelerrs = append(labelerrs, "Invalid label name - label=["+label+"%], labelname=["+labelname+"], errors: ")
			for _, err := range errs {
				labelerrs = append(labelerrs, err)
			}
		}
	}
	if len(labelvalue) > 0 {
		errs := validation.IsValidLabelValue(labelvalue)
		if len(errs) > 0 {
			labelerrs = append(labelerrs, "Invalid label value - label=["+label+"%], labelvalue=["+labelvalue+"], errors: ")
			for _, err := range errs {
				labelerrs = append(labelerrs, err)
			}
		}
	} else {
		// expect a non-zero value
		labelerrs = append(labelerrs, "Invalid label value - label=["+label+"%], labelvalue=["+labelvalue+"]")
	}

	return labelerrs
}

func IsValidNumber(value, min, max int) []string {
	var errs []string

	if min > max {
		errs = append(errs, "invalid constraints")
		return errs
	}

	if value < min {
		errs = append(errs, "value less than minimum")
	}
	if value > max {
		errs = append(errs, "value greater than maximum")
	}
	return errs
}

func IsValidNumberStr(value string, min, max int) []string {
	var errs []string

	if min > max {
		errs = append(errs, "invalid constraints")
		return errs
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		errs = append(errs, err.Error())
		return errs
	}
	if n < min {
		errs = append(errs, "value less than minimum")
	}
	if n > max {
		errs = append(errs, "value greater than maximum")
	}
	return errs
}

/*
IsValidParameterPresent method takes in a vars map and a array of string parameters
that you expect to be present in the GET request.
Returns Nil if all the parameters are present or else shall return error message.
*/
func IsValidParameterPresent(vars map[string]string, sp []string) error {

	for i := range sp {
		v := vars[sp[i]]
		if v == "" {
			errMessage := fmt.Sprintf("Missing %v in GET request", sp[i])
			return fmt.Errorf(errMessage)
		}

	}
	return nil

}

// ValidateJsonSchemaData function validates the document against the Json Schema
func ValidateJsonSchemaData(jsonSchemaFile string, jsonData interface{}) (error, int) {

	// Read the Json Schema File
	if _, err := os.Stat(jsonSchemaFile); err != nil {
		if os.IsNotExist(err) {
			err = pkgerrors.New("JsonSchemaValidation: File " + jsonSchemaFile + " not found")
		} else {
			err = pkgerrors.Wrap(err, "JsonSchemaValidation: Stat file error")
		}
		return err, http.StatusInternalServerError
	}
	rawBytes, err := ioutil.ReadFile(jsonSchemaFile)
	if err != nil {
		return pkgerrors.Wrap(err, "JsonSchemaValidation: Read JSON file error"), http.StatusInternalServerError
	}

	// Json encode the data
	req, err := json.Marshal(jsonData)
	if err != nil {
		return pkgerrors.Wrap(err, "JsonSchemaValidation, Request Body error"), http.StatusBadRequest
	}

	// Load schema and document
	schemaLoader := gojsonschema.NewStringLoader(string(rawBytes))
	s, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return pkgerrors.Wrap(err, "JsonSchemaValidation: Validation error"), http.StatusInternalServerError
	}
	documentLoader := gojsonschema.NewStringLoader(string(req))
	result, err := s.Validate(documentLoader)
	if err != nil {
		return pkgerrors.Wrap(err, "JsonSchemaValidation: Validation error"), http.StatusInternalServerError
	}
	// Validate document against Json Schema
	if !result.Valid() {
		for _, desc := range result.Errors() {
			log.Error("The document is not valid", log.Fields{
				"Error": desc.Description(),
			})
		}
		return pkgerrors.New("JsonSchemaValidation: Document Validation failed"), http.StatusBadRequest
	}

	return nil, 0
}
