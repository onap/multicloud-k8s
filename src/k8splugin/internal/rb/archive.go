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
	"io/ioutil"
	"os"
	"path/filepath"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
)

func isTarGz(r io.Reader) error {
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

//ExtractTarBall provides functionality to extract a tar.gz file
//into a temporary location for later use.
//It returns the path to the new location
func ExtractTarBall(r io.Reader) (string, error) {
	//Check if it is a valid gz
	gzf, err := gzip.NewReader(r)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Invalid gzip format")
	}

	//Check if it is a valid tar file
	//Unfortunately this can only be done by inspecting all the tar contents
	tarR := tar.NewReader(gzf)
	first := true

	outDir, _ := ioutil.TempDir("", "k8s-ext-")

	for true {
		header, err := tarR.Next()

		if err == io.EOF {
			//Check if we have just a gzip file without a tar archive inside
			if first {
				return "", pkgerrors.New("Empty or non-existant Tar file found")
			}
			//End of archive
			break
		}

		if err != nil {
			return "", pkgerrors.Wrap(err, "Error reading tar file")
		}

		target := filepath.Join(outDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				// Using 755 read, write, execute for owner
				// groups and others get read and execute permissions
				// on the folder.
				if err := os.MkdirAll(target, 0755); err != nil {
					return "", pkgerrors.Wrap(err, "Creating directory")
				}
			}
		case tar.TypeReg:
			err = utils.EnsureDirectory(target)
			if err != nil {
				return "", pkgerrors.Wrap(err, "Creating Directory")
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return "", pkgerrors.Wrap(err, "Creating file")
			}

			// copy over contents
			if _, err := io.Copy(f, tarR); err != nil {
				return "", pkgerrors.Wrap(err, "Copying file content")
			}

			// close for each file instead of a defer for all
			// at the end of the function
			f.Close()
		}

		first = false
	}

	return outDir, nil
}
