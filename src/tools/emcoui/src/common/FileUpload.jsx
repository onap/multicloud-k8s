//=======================================================================
// Copyright (c) 2017-2020 Aarna Networks, Inc.
// All rights reserved.
// ======================================================================
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//           http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ========================================================================
import React from "react";
import PropTypes from "prop-types";
import FileCopyIcon from "@material-ui/icons/FileCopy";
import CloudUploadIcon from "@material-ui/icons/CloudUpload";
import "./fileUpload.css";

const FileUpload = (props) => {
  return (
    <>
      <div className="file-upload">
        <div
          className="file-upload-wrap"
          style={{
            border:
              props.file &&
              props.file.name &&
              "2px dashed rgba(0, 131, 143, 1)",
          }}
        >
          <input
            required
            className="file-upload-input"
            type="file"
            accept={props.accept ? props.accept : "*"}
            name="file"
            onBlur={props.handleBlur ? props.handleBlur : null}
            onChange={(event) => {
              props.setFieldValue(props.name, event.currentTarget.files[0]);
            }}
          />

          <div className="file-upload-text">
            {props.file && props.file.name ? (
              <>
                <span>
                  <FileCopyIcon color="primary" />
                </span>
                <span style={{ fontWeight: 600 }}>{props.file.name}</span>
              </>
            ) : (
              <>
                <span>
                  <CloudUploadIcon />
                </span>
                <span>Drag And Drop or Click To Upload</span>
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
};

FileUpload.propTypes = {
  handleBlur: PropTypes.func,
  setFieldValue: PropTypes.func.isRequired,
};

export default FileUpload;
