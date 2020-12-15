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
import { Grid, Paper, Typography } from "@material-ui/core";
import EnhancedTable from "./SortableTable";

function AppFormPlacement({
  formikProps,
  index,
  clusterProviders,
  handleRowSelect,
  ...props
}) {
  return (
    <>
      <Typography variant="subtitle1" style={{ float: "left" }}>
        Select Clusters
        <span className="MuiFormLabel-asterisk MuiInputLabel-asterisk"> *</span>
      </Typography>
      {formikProps.errors.apps &&
        formikProps.errors.apps[index] &&
        formikProps.errors.apps[index].clusters && (
          <span
            style={{
              color: "#f44336",
              marginRight: "35px",
              float: "right",
            }}
          >
            {typeof formikProps.errors.apps[index].clusters === "string" &&
              formikProps.errors.apps[index].clusters}
          </span>
        )}
      <Grid
        container
        spacing={3}
        style={{
          height: "400px",
          overflowY: "auto",
          width: "100%",
          scrollbarWidth: "thin",
        }}
      >
        {clusterProviders &&
          clusterProviders.length > 0 &&
          clusterProviders.map((clusterProvider) => (
            <Grid key={clusterProvider.name} item xs={12}>
              <Paper>
                <EnhancedTable
                  key={clusterProvider.name}
                  tableName={clusterProvider.name}
                  clusters={clusterProvider.clusters}
                  formikValues={formikProps.values.apps[index].clusters}
                  onRowSelect={handleRowSelect}
                />
              </Paper>
            </Grid>
          ))}
      </Grid>
    </>
  );
}

AppFormPlacement.propTypes = {
  formikProps: PropTypes.object,
  index: PropTypes.number,
  handleRowSelect: PropTypes.func,
};

export default AppFormPlacement;
