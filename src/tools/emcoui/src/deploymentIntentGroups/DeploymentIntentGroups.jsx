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
import { withStyles, Button, Grid } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../services/apiService";
import Spinner from "../common/Spinner";
import DIGform from "./DIGform";

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
    let payload = {
      metadata: {
        name: inputFields.name,
        description: inputFields.description,
      },
      spec: {
        profile: inputFields.compositeProfile,
        version: inputFields.version,
      },
      projectName: props.projectName,
      compositeAppName: inputFields.compositeApp,
      compositeAppVersion: inputFields.compositeAppVersion,
    };
    if (inputFields.overrideValues && inputFields.overrideValues !== "") {
      payload.spec["override-values"] = JSON.parse(inputFields.overrideValues);
    }
    apiService
      .createDeploymentIntentGroup(payload)
      .then((response) => {
        response.compositeAppName = inputFields.compositeApp;
        response.compositeAppVersion = inputFields.compositeAppVersion;
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
  };

  useEffect(() => {
    apiService
      .getCompositeApps({ projectName: props.projectName })
      .then((response) => {
        const getDigIntents = (input) => {
          let request = {
            projectName: props.projectName,
            compositeAppName: input.compositeAppName,
            compositeAppVersion: input.compositeAppVersion,
            deploymentIntentGroupName: input.metadata.name,
          };
          apiService
            .getDeploymentIntentGroupIntents(request)
            .then((res) => {
              input.intent = res.intent;
            })
            .catch((err) => {})
            .finally(() => {
              setData((data) => [...data, input]);
            });
        };
        response.forEach((compositeApp) => {
          let request = {
            projectName: props.projectName,
            compositeAppName: compositeApp.metadata.name,
            compositeAppVersion: compositeApp.spec.version,
          };
          apiService
            .getDeploymentIntentGroups(request)
            .then((digResponse) => {
              digResponse.forEach((res) => {
                res.compositeAppName = compositeApp.metadata.name;
                res.compositeAppVersion = compositeApp.spec.version;
                getDigIntents(res);
              });
            })
            .catch((error) => {
              console.log("unable to get deployment intent groups", error);
            })
            .finally(() => {
              setCompositeApps(response);
              setIsloading(false);
            });
        });
      })
      .catch((err) => {
        console.log("Unable to get composite apps : ", err);
      });
  }, [props.projectName]);

  return (
    <>
      {isLoading && <Spinner />}
      {!isLoading && compositeApps && compositeApps.length > 0 && (
        <>
          <Button
            variant="outlined"
            color="primary"
            startIcon={<AddIcon />}
            onClick={onCreateDIG}
          >
            Create Deployment Intent Group
          </Button>
          <DIGform
            projectName={props.projectName}
            open={open}
            onClose={handleClose}
            onSubmit={handleSubmit}
            data={{ compositeApps: compositeApps }}
          />
          <Grid container spacing={2} alignItems="center">
            <Grid item xs style={{ marginTop: "20px" }}>
              <DIGtable
                data={data}
                setData={setData}
                projectName={props.projectName}
              />
            </Grid>
          </Grid>
        </>
      )}
    </>
  );
};
export default withStyles(styles)(DeploymentIntentGroups);
