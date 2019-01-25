/*
 * Copyright 2018 Intel Corporation, Inc
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

package rb

import (
	"archive/tar"
	"compress/gzip"
	pkgerrors "github.com/pkg/errors"
	"io"
)

func isTarGz(r io.Reader) error {
	//Check if it is a valid gz
	gzf, err := gzip.NewReader(r)
	if err != nil {
		return pkgerrors.Errorf("Invalid gz format %s", err.Error())
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
