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
import React, { useEffect, useState } from "react";
import DIGtable from "./DIGtable";
import { withStyles, Button, Grid, Typography } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../services/apiService";
import Spinner from "../common/Spinner";
import DIGform from "./DIGform";
import { ReactComponent as EmptyIcon } from "../assets/icons/empty.svg";

const styles = {
  root: {
    display: "flex",
    minHeight: "100vh",
  },
  app: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
  },
};

const DeploymentIntentGroups = (props) => {
  const [open, setOpen] = React.useState(false);
  const [data, setData] = useState([]);
  const [isLoading, setIsloading] = useState(true);
  const [compositeApps, setCompositeApps] = useState([]);
  const handleClose = () => {
    setOpen(false);
  };
  const onCreateDIG = () => {
    setOpen(true);
  };
  const handleSubmit = (inputFields) => {
    try {
      let payload = {
        spec: {
          projectName: props.projectName,
          appsData: inputFields.intents.apps,
        },
      };
      if (inputFields.overrideValues && inputFields.overrideValues !== "") {
        payload.spec["override-values"] = JSON.parse(
          inputFields.overrideValues
        );
      }
      payload = { ...payload, ...inputFields.general };
      apiService
        .createDeploymentIntentGroup(payload)
        .then((response) => {
          response.metadata.compositeAppName = inputFields.general.compositeApp;
          response.metadata.compositeAppVersion =
            inputFields.general.compositeAppVersion;
          data && data.length > 0
            ? setData([...data, response])
            : setData([response]);
        })
        .catch((error) => {
          console.log("error creating DIG : ", error);
        })
        .finally(() => {
          setIsloading(false);
          setOpen(false);
        });
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    let getDigs = () => {
      apiService
        .getDeploymentIntentGroups({ projectName: props.projectName })
        .then((res) => {
          setData(res);
        })
        .catch((err) => {
          console.log("error getting deplotment intent groups : " + err);
        })
        .finally(() => setIsloading(false));
    };

    apiService
      .getCompositeApps({ projectName: props.projectName })
      .then((response) => {
        setCompositeApps(response);
        getDigs();
      })
      .catch((err) => {
        console.log("Unable to get composite apps : ", err);
      });
  }, [props.projectName]);

  return (
    <>
      {isLoading && <Spinner />}
      {!isLoading && compositeApps && (
        <>
          <DIGform
            projectName={props.projectName}
            open={open}
            onClose={handleClose}
            onSubmit={handleSubmit}
            data={{ compositeApps: compositeApps }}
          />
          <Grid item xs={12}>
            <Button
              variant="outlined"
              color="primary"
              startIcon={<AddIcon />}
              onClick={onCreateDIG}
            >
              Create Deployment Intent Group
            </Button>
          </Grid>

          {data && data.length > 0 && (
            <Grid container spacing={2} alignItems="center">
              <Grid item xs style={{ marginTop: "20px" }}>
                <DIGtable
                  data={data}
                  setData={setData}
                  projectName={props.projectName}
                />
              </Grid>
            </Grid>
          )}

          {(data === null || (data && data.length < 1)) && (
            <Grid container spacing={2} direction="column" alignItems="center">
              <Grid style={{ marginTop: "60px" }} item xs={6}>
                <EmptyIcon style={{ height: "100px", width: "100px" }} />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="h6">
                  No deployment group found, start by adding a deployment group
                </Typography>
              </Grid>
            </Grid>
          )}
        </>
      )}
    </>
  );
};
export default withStyles(styles)(DeploymentIntentGroups);
