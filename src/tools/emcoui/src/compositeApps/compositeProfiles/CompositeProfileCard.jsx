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
import SupervisorAccountIcon from "@material-ui/icons/SupervisorAccount";
import AddIcon from "@material-ui/icons/Add";
import {
  TableContainer,
  Table,
  TableRow,
  TableHead,
  withStyles,
  Button,
} from "@material-ui/core";
import TableCell from "@material-ui/core/TableCell";
import Paper from "@material-ui/core/Paper";
import TableBody from "@material-ui/core/TableBody";
import apiService from "../../services/apiService";
// import EditIcon from "@material-ui/icons/Edit";
import DeleteIcon from "@material-ui/icons/Delete";
import ProfileForm from "./ProfileForm";
import DeleteDialog from "../../common/Dialogue";

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

export default function RecipeReviewCard(props) {
  const classes = useStyles();
  const [formOpen, setFormOpen] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [index, setIndex] = useState(0);
  const [data, setData] = useState([]);
  const handleExpandClick = () => {
    if (!expanded) {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        compositeProfileName: props.compositeProfile.metadata.name,
      };
      apiService
        .getProfiles(request)
        .then((res) => {
          setData(res);
        })
        .catch((err) => {
          console.log("error getting profiles : ", err);
        })
        .finally(setExpanded(!expanded));
    } else {
      setExpanded(!expanded);
    }
  };

  const handleEdit = () => {};
  const handleDelete = (index) => {
    setIndex(index);
    setOpenDialog(true);
  };

  const handleAddProfile = () => {
    setFormOpen(true);
  };
  const handleFormClose = () => {
    setFormOpen(false);
  };
  const handleCloseDialog = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        compositeProfileName: props.compositeProfile.metadata.name,
        profileName: data[index].metadata.name,
      };
      apiService
        .deleteProfile(request)
        .then(() => {
          console.log("profile deleted");
          data.splice(index, 1);
          setData([...data]);
        })
        .catch((err) => {
          console.log("Error deleting profile : ", err);
        })
        .finally(() => {
          setIndex(0);
        });
    }
    setOpenDialog(false);
  };

  const handleSubmit = (values) => {
    const formData = new FormData();
    formData.append("file", values.file);
    formData.append(
      "metadata",
      `{"metadata":{ "name": "${values.name}", "description": "${values.description}" }, "spec":{"app-name":"${values.appName}"}}`
    );
    formData.append("projectName", props.projectName);
    formData.append("compositeAppName", props.compositeAppName);
    formData.append("compositeAppVersion", props.compositeAppVersion);
    formData.append(
      "compositeProfileName",
      props.compositeProfile.metadata.name
    );
    apiService
      .addProfile(formData)
      .then((res) => {
        !data || data.length === 0 ? setData([res]) : setData([...data, res]);
      })
      .catch((err) => {
        console.log("error adding app : ", err);
      })
      .finally(() => {
        setFormOpen(false);
      });
  };
  return (
    <>
      {props.appsData && props.appsData.length > 0 && (
        <ProfileForm
          open={formOpen}
          onClose={handleFormClose}
          onSubmit={handleSubmit}
          appsData={props.appsData}
        />
      )}
      <Card className={classes.root}>
        <CardHeader
          onClick={handleExpandClick}
          avatar={<SupervisorAccountIcon fontSize="large" />}
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
          title={props.compositeProfile.metadata.name}
          subheader={props.compositeProfile.metadata.description}
        />
        <Collapse in={expanded} timeout="auto" unmountOnExit>
          <CardContent>
            {/* <Button
              disabled={!(props.appsData && props.appsData.length > 0)}
              variant="outlined"
              size="small"
              style={{ marginBottom: "15px" }}
              color="primary"
              startIcon={<AddIcon />}
              onClick={() => {
                handleAddProfile();
              }}
            >
              Add Profile
            </Button> */}
            {/* <Button
              variant="outlined"
              size="small"
              color="secondary"
              disabled={data && data.length > 0}
              style={{ float: "right" }}
              startIcon={<DeleteIcon />}
              onClick={() => {
                props.onDeleteCompositeProfile(props.index);
              }}
            >
              Delete Composite Profile
            </Button> */}
            {data && data.length > 0 && (
              <>
                <DeleteDialog
                  open={openDialog}
                  onClose={handleCloseDialog}
                  title={"Delete Profile"}
                  content={`Are you sure you want to delete "${
                    data[index] ? data[index].metadata.name : ""
                  }"`}
                />
                <TableContainer component={Paper}>
                  <Table className={classes.table}>
                    <TableHead>
                      <TableRow>
                        <StyledTableCell>Name</StyledTableCell>
                        <StyledTableCell>Description</StyledTableCell>
                        <StyledTableCell>App</StyledTableCell>
                        {/* <StyledTableCell>Actions</StyledTableCell> */}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {data.map((profile, index) => (
                        <StyledTableRow key={profile.metadata.name + index}>
                          <StyledTableCell>
                            {profile.metadata.name}
                          </StyledTableCell>
                          <StyledTableCell>
                            {profile.metadata.description}
                          </StyledTableCell>
                          <StyledTableCell>
                            {profile.spec["app-name"]}
                          </StyledTableCell>
                          {/* <StyledTableCell>
                            
                            //edit profile api is not implemented yet
                            <IconButton
                              onClick={(e) => handleEdit(index)}
                              title="Edit"
                            >
                              <EditIcon color="primary" />
                            </IconButton>
                            <IconButton
                              onClick={(e) => handleDelete(index)}
                              title="Delete"
                            >
                              <DeleteIcon color="secondary" />
                            </IconButton>
                          </StyledTableCell> */}
                        </StyledTableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </>
            )}
            {!(props.appsData && props.appsData.length > 0) && (
              <div>No app found for adding profile</div>
            )}
          </CardContent>
        </Collapse>
      </Card>
    </>
  );
}
