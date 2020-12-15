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
import React, { useState } from "react";
import { makeStyles } from "@material-ui/core/styles";
import clsx from "clsx";
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Collapse from "@material-ui/core/Collapse";
import IconButton from "@material-ui/core/IconButton";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import StorageIcon from "@material-ui/icons/Storage";
import AddIcon from "@material-ui/icons/Add";
import { Button } from "@material-ui/core";
import apiService from "../../services/apiService";
import AppPlacementIntentsTable from "./AppPlacementIntentTable";
import AppIntentForm from "./AppIntentForm";
import DeleteIcon from "@material-ui/icons/Delete";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
    marginBottom: "15px",
    boxShadow:
      "0px 3px 5px -1px rgba(0,0,0,0.2),0px 5px 8px 0px rgba(0,0,0,0.14),0px 1px 14px 0px rgba(0,0,0,0.12)",
  },
  expand: {
    transform: "rotate(0deg)",
    marginLeft: "auto",
    transition: theme.transitions.create("transform", {
      duration: theme.transitions.duration.shortest,
    }),
  },
  expandOpen: {
    transform: "rotate(180deg)",
  },
}));
const GenericPlacementIntentCard = (props) => {
  const classes = useStyles();
  const [formOpen, setFormOpen] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [appPlacementIntentData, setAppPlacementIntentData] = useState({});

  const handleExpandClick = () => {
    if (!expanded && !appPlacementIntentData.applications) {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        genericPlacementIntentName: props.genericPlacementIntent.metadata.name,
      };
      apiService
        .getAppPlacementIntents(request)
        .then((res) => {
          setAppPlacementIntentData(res);
        })
        .catch((err) => {
          console.log("error getting workload intents : ", err);
        })
        .finally(() => {
          setExpanded(!expanded);
        });
    } else {
      setExpanded(!expanded);
    }
  };
  const handleAddAppIntent = () => {
    setFormOpen(true);
  };
  const handleSubmit = (values) => {
    setFormOpen(false);
    let Intent = JSON.parse(values.intent);
    let request = {
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
      genericPlacementIntentName: props.genericPlacementIntent.metadata.name,
      payload: {
        metadata: { name: values.name, description: values.description },
        spec: { "app-name": values.appName, intent: Intent },
      },
    };
    apiService
      .addAppPlacementIntent(request)
      .then((res) => {
        let newData = {
          name: res.spec["app-name"],
          description: res.metadata.description,
          allOf: res.spec.intent.allOf,
        };
        console.log("app intent added to generic placement intent : ", newData);
        !appPlacementIntentData.applications ||
        appPlacementIntentData.applications.length < 1
          ? setAppPlacementIntentData({ applications: [newData] })
          : setAppPlacementIntentData((appPlacementIntentData) => {
              return {
                applications: [...appPlacementIntentData.applications, newData],
              };
            });
      })
      .catch((err) => {
        console.log(
          "unable to add app intent to generic placement intent : ",
          err
        );
      });
  };
  const handleFormClose = () => {
    setFormOpen(false);
  };

  return (
    <>
      {props.appsData && props.appsData.length > 0 && (
        <AppIntentForm
          open={formOpen}
          onClose={handleFormClose}
          onSubmit={handleSubmit}
          appsData={props.appsData}
        />
      )}
      <Card className={classes.root}>
        <CardHeader
          onClick={handleExpandClick}
          avatar={<StorageIcon fontSize="large" />}
          action={
            <IconButton
              className={clsx(classes.expand, {
                [classes.expandOpen]: expanded,
              })}
              onClick={handleExpandClick}
              aria-expanded={expanded}
            >
              <ExpandMoreIcon />
            </IconButton>
          }
          title={props.genericPlacementIntent.metadata.name}
          subheader={props.genericPlacementIntent.metadata.description}
        />
        <Collapse in={expanded} timeout="auto" unmountOnExit>
          <CardContent>
            <Button
              disabled={!(props.appsData && props.appsData.length > 0)}
              variant="outlined"
              size="small"
              style={{ marginBottom: "15px" }}
              color="primary"
              startIcon={<AddIcon />}
              onClick={() => {
                handleAddAppIntent();
              }}
            >
              Add App Placement Intent
            </Button>
            <Button
              variant="outlined"
              size="small"
              color="secondary"
              disabled={
                appPlacementIntentData.applications &&
                appPlacementIntentData.applications.length > 0
              }
              style={{ float: "right" }}
              startIcon={<DeleteIcon />}
              onClick={() => {
                props.onDeleteGenericPlacementIntent(props.index);
              }}
            >
              Delete Placement Intent
            </Button>
            {appPlacementIntentData.applications &&
              appPlacementIntentData.applications.length > 0 && (
                <AppPlacementIntentsTable
                  data={appPlacementIntentData.applications}
                  setData={setAppPlacementIntentData}
                  appsData={props.appsData}
                  projectName={props.projectName}
                  compositeAppName={props.compositeAppName}
                  compositeAppVersion={props.compositeAppVersion}
                  genericPlacementIntentName={
                    props.genericPlacementIntent.metadata.name
                  }
                />
              )}
            {!(props.appsData && props.appsData.length > 0) && (
              <div>No app found for adding app placement intent</div>
            )}
          </CardContent>
        </Collapse>
      </Card>
    </>
  );
};

GenericPlacementIntentCard.propTypes = {};

export default GenericPlacementIntentCard;
