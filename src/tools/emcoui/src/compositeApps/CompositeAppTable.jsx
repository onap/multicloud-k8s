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
// import { Link } from "react-router-dom";
import IconButton from "@material-ui/core/IconButton";
import EditIcon from "@material-ui/icons/Edit";
import DeleteIcon from "@material-ui/icons/Delete";
import CreateCompositeAppForm from "./dialogs/CompositeAppForm";
import apiService from "../services/apiService";
import DeleteDialog from "../common/Dialogue";
import { Link, withRouter } from "react-router-dom";
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

function CustomizedTables({ data, ...props }) {
  const classes = useStyles();
  const [openForm, setOpenForm] = useState(false);
  const [activeRowIndex, setActiveRowIndex] = useState(0);
  const [row, setRow] = useState({});
  const [open, setOpen] = useState(false);
  const [notificationDetails, setNotificationDetails] = useState({});
  let onEditCompositeApp = (row, index) => {
    setActiveRowIndex(index);
    setRow(row);
    setOpenForm(true);
    // props.history.push(`services/${row.metadata.name}/${row.spec.version}`);
  };
  const handleCloseForm = (fields) => {
    if (fields) {
      let request = {
        payload: {
          name: fields.name,
          description: fields.description,
          spec: { version: fields.version },
        },
        projectName: props.projectName,
        compositeAppVersion: row.spec.version,
      };
      apiService
        .updateCompositeApp(request)
        .then((res) => {
          let updatedData = data.slice();
          updatedData.splice(activeRowIndex, 1);
          updatedData.push(res);
          props.handleUpdateState(updatedData);
        })
        .catch((err) => {
          console.log("error creating composite app : ", err);
        })
        .finally(() => {
          setOpenForm(false);
        });
    } else {
      setOpenForm(false);
    }
  };
  const handleDeleteCompositeApp = (index) => {
    setActiveRowIndex(index);
    setOpen(true);
  };
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: data[activeRowIndex].metadata.name,
        compositeAppVersion: data[activeRowIndex].spec.version,
      };
      apiService
        .deleteCompositeApp(request)
        .then(() => {
          console.log("cluster deleted");
          data.splice(activeRowIndex, 1);
          let updatedData = data.slice();
          props.handleUpdateState(updatedData);
        })
        .catch((err) => {
          console.log("Error deleting cluster : ", err);
          let message = "Error deleting service";
          if (err.response.data.includes("Non emtpy DIG in service")) {
            message =
              "Error deleting service : please delete deployment intent group first";
          }
          setNotificationDetails({
            show: true,
            message: message,
            severity: "error",
          });
        });
    }
    setOpen(false);
    setActiveRowIndex(0);
  };

  return (
    <>
      <Notification notificationDetails={notificationDetails} />
      {data && data.length > 0 && (
        <>
          <CreateCompositeAppForm
            open={openForm}
            handleClose={handleCloseForm}
            item={row}
          />
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Delete Service"}
            content={`Are you sure you want to delete "${
              data[activeRowIndex] ? data[activeRowIndex].metadata.name : ""
            }" ?`}
          />
          <TableContainer component={Paper}>
            <Table className={classes.table} size="small">
              <TableHead>
                <TableRow>
                  <StyledTableCell>Name</StyledTableCell>
                  <StyledTableCell>Description</StyledTableCell>
                  <StyledTableCell>Version</StyledTableCell>
                  <StyledTableCell>Actions</StyledTableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {data.map((row, index) => (
                  <StyledTableRow key={row.metadata.name}>
                    <StyledTableCell>
                      <Link
                        to={`services/${row.metadata.name}/${row.spec.version}`}
                      >
                        {row.metadata.name}
                      </Link>
                      {/* {row.metadata.name} */}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.metadata.description}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.version}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {/* <IconButton
                        onClick={(e) => onEditCompositeApp(row, index)}
                        title="Edit"
                      >
                        <EditIcon color="primary" />
                      </IconButton> */}
                      <IconButton
                        color="secondary"
                        onClick={() => {
                          handleDeleteCompositeApp(index);
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </StyledTableCell>
                  </StyledTableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </>
      )}
      {(!data || data.length === 0) && <span>No Composite Apps</span>}
    </>
  );
}

export default withRouter(CustomizedTables);
