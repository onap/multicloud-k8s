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
import DeleteDialog from "../../common/Dialogue";
import DeleteIcon from "@material-ui/icons/Delete";
// import ControllerForm from "./ControllerForm";
import apiService from "../../services/apiService";

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

export default function ControllersTable(props) {
  const classes = useStyles();
  const [open, setOpen] = React.useState(false);
  // const [openForm, setOpenForm] = React.useState(false);
  const [index, setIndex] = React.useState(0);

  // commenting below code as edit is currently not supported

  // let handleEdit = (index) => {
  //   setIndex(index);
  //   setOpenForm(true);
  // };
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      apiService
        .removeController(props.data[index].metadata.name)
        .then(() => {
          console.log("controller removed");
          props.data.splice(index, 1);
          props.setControllerData([...props.data]);
        })
        .catch((err) => {
          console.log("Error removing controller : ", err);
        });
    }
    setOpen(false);
    setIndex(0);
  };
  // commenting below code as edit is not currently supported
  // const handleFormClose = (values) => {
  //   if (values) {
  //     let request = {
  //       metadata: { name: values.name, description: values.description },
  //       spec: {
  //         host: values.host,
  //         port: parseInt(values.port),
  //       },
  //     };
  //     if (values.type) request.spec.type = values.type;
  //     if (values.priority) request.spec.priority = parseInt(values.priority);
  //     apiService
  //       .updateController(request)
  //       .then((res) => {
  //         props.data[index] = res;
  //         props.setControllersData([...props.data]);
  //       })
  //       .catch((err) => {
  //         console.log("error adding controller : " + err);
  //       });
  //   }
  //   setOpenForm(false);
  // };
  const handleDelete = (index) => {
    setIndex(index);
    setOpen(true);
  };

  // commenting below codecd as edit is not currently supported
  // const handleSubmit = (data) => {
  //   let payload = { metadata: data };
  //   apiService
  //     .updateProject(payload)
  //     .then((res) => {
  //       props.data[index] = res;
  //       props.setProjectsData([...props.data]);
  //     })
  //     .catch((err) => {
  //       console.log("Error updating project : ", err);
  //     });
  //   setOpenForm(false);
  // };

  return (
    <React.Fragment>
      {props.data && props.data.length > 0 && (
        <>
          {/* 
          //commenting edit as edit is not currently supported
          <ControllerForm
            open={openForm}
            onClose={handleFormClose}
            item={props.data[index]}
            onSubmit={handleSubmit}
          /> */}
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Remove Controller"}
            content={`Are you sure you want to delete "${
              props.data[index] ? props.data[index].metadata.name : ""
            }" ?`}
          />
          <TableContainer component={Paper}>
            <Table className={classes.table} size="small">
              <TableHead>
                <TableRow>
                  <StyledTableCell>Name</StyledTableCell>
                  <StyledTableCell>Description</StyledTableCell>
                  <StyledTableCell>Host</StyledTableCell>
                  <StyledTableCell>Port</StyledTableCell>
                  <StyledTableCell>Type</StyledTableCell>
                  <StyledTableCell>Priority</StyledTableCell>
                  <StyledTableCell>Actions</StyledTableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {props.data.map((row, index) => (
                  <StyledTableRow key={row.metadata.name + "" + index}>
                    <StyledTableCell className={classes.cell}>
                      {row.metadata.name}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.metadata.description}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.host}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.port}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.type}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.spec.priority}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {/* 
                      //commenting edit as edit is not currently supported
                      <IconButton
                        onClick={(e) => handleEdit(index)}
                        title="Edit"
                      >
                        <EditIcon color="primary" />
                      </IconButton> */}
                      <IconButton
                        onClick={(e) => handleDelete(index)}
                        title="Remove"
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
