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
import { withStyles, makeStyles } from "@material-ui/core/styles";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import Paper from "@material-ui/core/Paper";
import IconButton from "@material-ui/core/IconButton";
// import EditIcon from "@material-ui/icons/Edit";
import DeleteDialog from "../common/Dialogue";
// import AddIcon from "@material-ui/icons/Add";
import DeleteIcon from "@material-ui/icons/Delete";
import GetAppIcon from "@material-ui/icons/GetApp";
import apiService from "../services/apiService";
// import { Button } from "@material-ui/core";
// import IntentsForm from "./IntentsForm";
import Notification from "../common/Notification";

const StyledTableCell = withStyles((theme) => ({
  body: {
    fontSize: 14,
  },
}))(TableCell);

const StyledTableRow = withStyles((theme) => ({
  root: {
    "&:nth-of-type(odd)": {
      backgroundColor: theme.palette.action.hover,
    },
  },
}))(TableRow);

const useStyles = makeStyles({
  table: {
    minWidth: 350,
  },
  cell: {
    color: "grey",
  },
});

export default function DIGtable({ data, setData, ...props }) {
  const classes = useStyles();
  const [open, setOpen] = useState(false);
  const [index, setIndex] = useState(0);
  // const [openIntentsForm, setOpenIntentsForm] = useState(false);
  const [notificationDetails, setNotificationDetails] = useState({});
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: data[index].metadata.compositeAppName,
        compositeAppVersion: data[index].metadata.compositeAppVersion,
        deploymentIntentGroupName: data[index].metadata.name,
      };
      apiService
        .deleteDeploymentIntentGroup(request)
        .then(() => {
          console.log("DIG deleted");
          data.splice(index, 1);
          setData([...data]);
        })
        .catch((err) => {
          console.log("Error deleting DIG : ", err);
        });
    }
    setOpen(false);
    setIndex(0);
  };
  const handleDelete = (index) => {
    setIndex(index);
    setOpen(true);
  };
  // const handleCloseIntentsForm = () => {
  //   setOpenIntentsForm(false);
  // };
  // const handleSubmitIntentForm = (values) => {
  //   setOpenIntentsForm(false);
  //   let request = {
  //     projectName: props.projectName,
  //     compositeAppName: values.compositeAppName,
  //     compositeAppVersion: values.compositeAppVersion,
  //     deploymentIntentGroupName: values.deploymentIntentGroupName,
  //     payload: {
  //       metadata: { name: values.name, description: values.description },
  //       spec: {
  //         intent: {
  //           genericPlacementIntent: values.genericPlacementIntent,
  //         },
  //       },
  //     },
  //   };
  //   if (values.networkControllerIntent && values.networkControllerIntent !== "")
  //     request.payload.spec.intent.ovnaction = values.networkControllerIntent;
  //   apiService
  //     .addIntentsToDeploymentIntentGroup(request)
  //     .then((res) => {
  //       if (data[index].intent) {
  //         data[index].intent.push(res.spec.intent);
  //       } else {
  //         data[index].intent = [res.spec.intent];
  //       }
  //       setData([...data]);
  //     })
  //     .catch((err) => {
  //       console.log("error adding intent to deployment intent group");
  //     });
  // };
  const handleInstantiate = (index) => {
    let request = {
      projectName: props.projectName,
      compositeAppName: data[index].metadata.compositeAppName,
      compositeAppVersion: data[index].metadata.compositeAppVersion,
      deploymentIntentGroupName: data[index].metadata.name,
    };
    apiService
      .approveDeploymentIntentGroup(request)
      .then(() => {
        console.log(
          "Deployment intent group approved, now going to instantiate"
        );
        apiService
          .instantiate(request)
          .then((res) => {
            console.log("Deployment intent group instantiated : " + res);
            setNotificationDetails({
              show: true,
              message: `Deployment intent group "${data[index].metadata.name}" instantiated`,
              severity: "success",
            });
          })
          .catch((err) => {
            console.log(
              `Error instantiating "${data[index].metadata.name}" deployment intent group: ` +
                err
            );
            setNotificationDetails({
              show: true,
              message: `Error instantiating "${data[index].metadata.name}" deployment intent group`,
              severity: "error",
            });
          });
      })
      .catch((err) => {
        console.log(
          `Error approving "${data[index].metadata.name}" deployment intent group : ` +
            err
        );
        setNotificationDetails({
          show: true,
          message: `Error approving "${data[index].metadata.name}" deployment intent group`,
          severity: "error",
        });
      });
  };

  return (
    <React.Fragment>
      <Notification notificationDetails={notificationDetails} />
      {data && data.length > 0 && (
        <>
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Delete Deployment Intent Group"}
            content={`Are you sure you want to delete "${
              data[index] ? data[index].metadata.name : ""
            }" ?`}
          />
          <TableContainer component={Paper}>
            <Table className={classes.table} size="small">
              <TableHead>
                <TableRow>
                  <StyledTableCell>Name</StyledTableCell>
                  <StyledTableCell>Version</StyledTableCell>
                  <StyledTableCell>Profile</StyledTableCell>
                  <StyledTableCell>Composite App</StyledTableCell>
                  {/* <StyledTableCell>Intents</StyledTableCell> */}
                  <StyledTableCell>Description</StyledTableCell>
                  <StyledTableCell style={{ width: "15%" }}>
                    Actions
                  </StyledTableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {data.map((row, index) => (
                  <StyledTableRow key={row.metadata.name + "" + index}>
                    <StyledTableCell>{row.metadata.name}</StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.version}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.profile}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.metadata.compositeAppName}
                    </StyledTableCell>
                    {/* {
                      <StyledTableCell className={classes.cell}>
                        {Object.keys(row.spec.deployedIntents[0]).map(function (
                          key,
                          index
                        ) {
                          if (
                            index === 0 ||
                            row.spec.deployedIntents[0][key] === ""
                          )
                            return row.spec.deployedIntents[0][key];
                          else return ", " + row.spec.deployedIntents[0][key];
                        })}
                      </StyledTableCell>
                    } */}
                    <StyledTableCell className={classes.cell}>
                      {row.metadata.description}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      <IconButton
                        color={"primary"}
                        // disabled={
                        //   !(
                        //     row.spec.deployedIntents &&
                        //     row.spec.deployedIntents.length > 0
                        //   )
                        // }
                        title="Instantiate"
                        onClick={(e) => handleInstantiate(index)}
                      >
                        <GetAppIcon />
                      </IconButton>
                      <IconButton
                        onClick={(e) => handleDelete(index)}
                        title="Delete"
                      >
                        <DeleteIcon color="secondary" />
                      </IconButton>
                    </StyledTableCell>
                  </StyledTableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </>
      )}
    </React.Fragment>
  );
}
