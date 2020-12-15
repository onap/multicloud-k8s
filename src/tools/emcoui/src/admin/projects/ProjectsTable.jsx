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
import { Link } from "react-router-dom";
import IconButton from "@material-ui/core/IconButton";
import EditIcon from "@material-ui/icons/Edit";
import DeleteDialog from "../../common/Dialogue";
import DeleteIcon from "@material-ui/icons/Delete";
import ProjectForm from "./ProjectForm";
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

export default function ProjectsTable(props) {
  const classes = useStyles();
  const [open, setOpen] = React.useState(false);
  const [openForm, setOpenForm] = React.useState(false);
  const [index, setIndex] = React.useState(0);

  let handleEdit = (index) => {
    setIndex(index);
    setOpenForm(true);
  };
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      apiService
        .deleteProject(props.data[index].metadata.name)
        .then(() => {
          console.log("project deleted");
          props.data.splice(index, 1);
          props.setProjectsData([...props.data]);
        })
        .catch((err) => {
          console.log("Error deleting project : ", err);
        });
    }
    setOpen(false);
    setIndex(0);
  };
  const handleFormClose = () => {
    setIndex(0);
    setOpenForm(false);
  };
  const handleDelete = (index) => {
    setIndex(index);
    setOpen(true);
  };
  const handleSubmit = (data) => {
    let payload = { metadata: data };
    apiService
      .updateProject(payload)
      .then((res) => {
        props.data[index] = res;
        props.setProjectsData([...props.data]);
      })
      .catch((err) => {
        console.log("Error updating project : ", err);
      });
    setOpenForm(false);
  };

  return (
    <React.Fragment>
      {props.data && props.data.length > 0 && (
        <>
          <ProjectForm
            open={openForm}
            onClose={handleFormClose}
            item={props.data[index]}
            onSubmit={handleSubmit}
          />
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Delete Project"}
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
                  <StyledTableCell>Actions</StyledTableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {props.data.map((row, index) => (
                  <StyledTableRow key={row.metadata.name + "" + index}>
                    <StyledTableCell>
                      {" "}
                      <Link to={`/app/projects/${row.metadata.name}`}>
                        {row.metadata.name}
                      </Link>
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
                      {row.metadata.description}
                    </StyledTableCell>
                    <StyledTableCell className={classes.cell}>
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
