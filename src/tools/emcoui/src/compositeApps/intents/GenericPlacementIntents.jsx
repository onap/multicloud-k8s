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
import { Button, Grid, makeStyles } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import PropTypes from "prop-types";
import Form from "./GenericPlacementIntentForm";
import apiService from "../../services/apiService";
import GenericPlacementIntentCard from "./GenericPlacementIntentCard";
import DeleteDialog from "../../common/Dialogue";
import Notification from "../../common/Notification";

const useStyles = makeStyles((theme) => ({
  root: {
    display: "flex",
    marginTop: "15px",
  },
  details: {
    display: "flex",
    flexDirection: "column",
  },
  content: {
    flex: "1 0 auto",
  },
  cover: {
    width: 151,
  },
  cardRoot: {
    boxShadow:
      "0px 3px 5px -1px rgba(0,0,0,0.2),0px 5px 8px 0px rgba(0,0,0,0.14),0px 1px 14px 0px rgba(0,0,0,0.12)",
  },
}));
const CompositeProfiles = (props) => {
  const classes = useStyles();
  const [isLoading, setIsLoading] = useState(true);
  const [openForm, setOpenForm] = useState(false);
  const [data, setData] = useState([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [index, setIndex] = useState(0);
  const [notificationDetails, setNotificationDetails] = useState({
    show: false,
    severity: "success",
    message: "",
  });

  useEffect(() => {
    let request = {
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
    };
    apiService
      .getGenericPlacementIntents(request)
      .then((res) => {
        setData(res);
      })
      .catch((err) => {
        console.log("error geting composite profiles ", err);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [props.projectName, props.compositeAppName, props.compositeAppVersion]);
  const handleCloseForm = () => {
    setOpenForm(false);
  };
  const handleAddCompositeProfile = () => {
    setOpenForm(true);
  };
  const handleSubmit = (values) => {
    let spec = values.spec ? JSON.parse(values.spec) : "";
    let request = {
      payload: {
        metadata: { name: values.name, description: values.description },
        spec: spec,
      },
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
    };
    apiService
      .createGenericPlacementIntent(request)
      .then((res) => {
        !data || data.length === 0 ? setData([res]) : setData([...data, res]);
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
        genericPlacementIntentName: data[index].metadata.name,
      };
      apiService
        .deleteGenericPlacementIntent(request)
        .then(() => {
          console.log("generic placement intent deleted");
          data.splice(index, 1);
          setData([...data]);
        })
        .catch((err) => {
          console.log("Error deleting generic placement intent : " + err);
          setNotificationDetails({
            show: true,
            message: "Error deleting generic placement intent ",
            severity: "error",
          });
        })
        .finally(() => {
          setIndex(0);
        });
    }
    setOpenDialog(false);
  };
  const handleDeleteGenericPlacementIntent = (index) => {
    setIndex(index);
    setOpenDialog(true);
  };
  return (
    <>
      <Notification notificationDetails={notificationDetails} />
      <DeleteDialog
        open={openDialog}
        onClose={handleCloseDialog}
        title={"Delete App Placement Intent"}
        content={`Are you sure you want to delete "${
          data && data[index] ? data[index].metadata.name : ""
        }"`}
      />
      <Button
        disabled={isLoading}
        variant="outlined"
        color="primary"
        startIcon={<AddIcon />}
        onClick={handleAddCompositeProfile}
      >
        Create Generic Placement Intent
      </Button>
      <Form onClose={handleCloseForm} open={openForm} onSubmit={handleSubmit} />
      <Grid container justify="flex-start" className={classes.root}>
        {data &&
          data.map((genericPlacementIntent, index) => (
            <GenericPlacementIntentCard
              key={genericPlacementIntent.metadata.name}
              genericPlacementIntent={genericPlacementIntent}
              projectName={props.projectName}
              compositeAppName={props.compositeAppName}
              compositeAppVersion={props.compositeAppVersion}
              appsData={props.appsData}
              index={index}
              onDeleteGenericPlacementIntent={
                handleDeleteGenericPlacementIntent
              }
            />
          ))}
      </Grid>
    </>
  );
};

CompositeProfiles.propTypes = {
  projectName: PropTypes.string.isRequired,
  compositeAppName: PropTypes.string.isRequired,
};

export default CompositeProfiles;
