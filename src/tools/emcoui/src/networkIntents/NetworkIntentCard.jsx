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
import SettingsEthernetIcon from "@material-ui/icons/SettingsEthernet";
import AddIcon from "@material-ui/icons/Add";
import { Button } from "@material-ui/core";
import apiService from "../services/apiService";
import WorkloadIntentTable from "./WorkloadIntentTable";
import DeleteIcon from "@material-ui/icons/Delete";
import WorkloadIntentForm from "./WorkloadIntentForm";

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

const NetworkIntentCard = (props) => {
  const classes = useStyles();
  const [formOpen, setFormOpen] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [workloadData, setWorkloadData] = useState([]);
  const handleExpandClick = () => {
    if (!expanded && workloadData && workloadData.length < 1) {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        networkControllerIntentName:
          props.networkControllerIntent.metadata.name,
      };
      apiService
        .getWorkloadIntents(request)
        .then((res) => {
          getInterfaces(res, expanded);
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
  const getInterfaces = (workloadIntentData) => {
    if (workloadIntentData && workloadIntentData.length > 0) {
      workloadIntentData.forEach((wokloadIntent) => {
        let request = {
          projectName: props.projectName,
          compositeAppName: props.compositeAppName,
          compositeAppVersion: props.compositeAppVersion,
          networkControllerIntentName:
            props.networkControllerIntent.metadata.name,
          workloadIntentName: wokloadIntent.metadata.name,
        };
        apiService
          .getInterfaces(request)
          .then((res) => {
            wokloadIntent.interfaces = res;
          })
          .catch((err) => {
            console.log("error getting workload intent interfaces : ", err);
          })
          .finally(() => {
            setWorkloadData([...workloadIntentData]);
          });
      });
    } else {
      setWorkloadData(workloadIntentData);
    }
  };
  const handleAddNetworkIntent = () => {
    setFormOpen(true);
  };
  const handleFormClose = () => {
    setFormOpen(false);
  };
  const handleSubmit = (values) => {
    const request = {
      payload: {
        metadata: {
          name: values.name,
          description: values.description,
        },
        spec: {
          "application-name": values.appName,
          "workload-resource": values.workloadResource,
          type: values.type,
        },
      },
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
      networkControllerIntentName: props.networkControllerIntent.metadata.name,
    };
    apiService
      .addWorkloadIntent(request)
      .then((res) => {
        !workloadData || workloadData.length === 0
          ? setWorkloadData([res])
          : setWorkloadData([...workloadData, res]);
      })
      .catch((err) => {
        console.log("error adding workload intent : ", err);
      })
      .finally(() => {
        setFormOpen(false);
      });
  };
  return (
    <>
      {props.appsData && props.appsData.length > 0 && (
        <WorkloadIntentForm
          open={formOpen}
          onClose={handleFormClose}
          onSubmit={handleSubmit}
          appsData={props.appsData}
        />
      )}
      <Card className={classes.root}>
        <CardHeader
          onClick={handleExpandClick}
          avatar={<SettingsEthernetIcon fontSize="large" />}
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
          title={props.networkControllerIntent.metadata.name}
          subheader={props.networkControllerIntent.metadata.description}
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
                handleAddNetworkIntent();
              }}
            >
              Add Workload Intent
            </Button>
            <Button
              variant="outlined"
              size="small"
              color="secondary"
              disabled={workloadData && workloadData.length > 0}
              style={{ float: "right" }}
              startIcon={<DeleteIcon />}
              onClick={props.onDeleteNetworkControllerIntent.bind(
                this,
                props.index
              )}
            >
              Delete Network Intent
            </Button>
            {workloadData && workloadData.length > 0 && (
              <WorkloadIntentTable
                data={workloadData}
                setData={setWorkloadData}
                projectName={props.projectName}
                compositeAppName={props.compositeAppName}
                compositeAppVersion={props.compositeAppVersion}
                networkControllerIntentName={
                  props.networkControllerIntent.metadata.name
                }
              />
            )}
            {!(props.appsData && props.appsData.length > 0) && (
              <div>No app found for adding workload intent</div>
            )}
          </CardContent>
        </Collapse>
      </Card>
    </>
  );
};

NetworkIntentCard.propTypes = {};
export default NetworkIntentCard;
