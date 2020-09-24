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
import React, { useState, useEffect } from "react";
import { Button, Grid } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../services/apiService";
import Card from "./NetworkIntentCard";
import Form from "../common/Form";
import DeleteDialog from "../common/Dialogue";

const NetworkIntents = (props) => {
  const [
    netwotkControllerIntentData,
    setNetwotkControllerIntentData,
  ] = useState([]);
  const [isLoading, setIsloading] = useState(true);
  const [openForm, setOpenForm] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [index, setIndex] = useState(0);
  useEffect(() => {
    let request = {
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
    };
    apiService
      .getNetworkControllerIntents(request)
      .then((res) => {
        setNetwotkControllerIntentData(res);
      })
      .catch((err) => {})
      .finally(() => {
        setIsloading(false);
      });
  }, [props.projectName, props.compositeAppName, props.compositeAppVersion]);

  const handleAddNetworkControllerIntent = () => {
    setOpenForm(true);
  };
  const handleCloseForm = () => {
    setOpenForm(false);
  };

  const handleSubmit = (values) => {
    let request = {
      payload: {
        metadata: { name: values.name, description: values.description },
      },
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
    };
    apiService
      .addNetworkControllerIntent(request)
      .then((res) => {
        !netwotkControllerIntentData || netwotkControllerIntentData.length === 0
          ? setNetwotkControllerIntentData([res])
          : setNetwotkControllerIntentData([
              ...netwotkControllerIntentData,
              res,
            ]);
      })
      .catch((err) => {
        console.log("error creating composite profile : ", err);
      })
      .finally(() => {
        setOpenForm(false);
      });
  };
  const handleCloseDialog = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        networkControllerIntentName:
          netwotkControllerIntentData[index].metadata.name,
      };
      apiService
        .deleteNetworkControllerIntent(request)
        .then(() => {
          console.log("generic placement intent deleted");
          netwotkControllerIntentData.splice(index, 1);
          setNetwotkControllerIntentData([...netwotkControllerIntentData]);
        })
        .catch((err) => {
          console.log("Error deleting generic placement intent : ", err);
        })
        .finally(() => {
          setIndex(0);
        });
    }
    setOpenDialog(false);
  };
  const handleDeleteNetworkControllerIntent = (index) => {
    setIndex(index);
    setOpenDialog(true);
  };
  return (
    <>
      <DeleteDialog
        open={openDialog}
        onClose={handleCloseDialog}
        title={"Delete Network Controller Intent"}
        content={`Are you sure you want to delete "${
          netwotkControllerIntentData && netwotkControllerIntentData[index]
            ? netwotkControllerIntentData[index].metadata.name
            : ""
        }"`}
      />

      <Form onClose={handleCloseForm} open={openForm} onSubmit={handleSubmit} />
      <Button
        disabled={isLoading}
        variant="outlined"
        color="primary"
        startIcon={<AddIcon />}
        onClick={handleAddNetworkControllerIntent}
      >
        Create Network Controller Intent
      </Button>
      <Grid
        container
        justify="flex-start"
        style={{ display: "flex", marginTop: "15px" }}
      >
        {netwotkControllerIntentData &&
          netwotkControllerIntentData.map(
            (networkControllerIntent, itemIndex) => (
              <Card
                key={networkControllerIntent.metadata.name}
                networkControllerIntent={networkControllerIntent}
                projectName={props.projectName}
                compositeAppName={props.compositeAppName}
                compositeAppVersion={props.compositeAppVersion}
                appsData={props.appsData}
                index={itemIndex}
                onDeleteNetworkControllerIntent={
                  handleDeleteNetworkControllerIntent
                }
              />
            )
          )}
      </Grid>
    </>
  );
};

NetworkIntents.propTypes = {};
export default NetworkIntents;
