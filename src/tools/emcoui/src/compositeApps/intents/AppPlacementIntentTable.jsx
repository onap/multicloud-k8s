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
import {
  TableContainer,
  Table,
  TableRow,
  TableHead,
  withStyles,
  TableCell,
  Chip,
} from "@material-ui/core";
import Paper from "@material-ui/core/Paper";
import TableBody from "@material-ui/core/TableBody";
// import EditIcon from "@material-ui/icons/Edit";
import DeleteIcon from "@material-ui/icons/Delete";
import PropTypes from "prop-types";
import apiService from "../../services/apiService";
import IconButton from "@material-ui/core/IconButton";
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

const AppPlacementIntentTable = ({ data, setData, ...props }) => {
  const [index, setIndex] = useState(0);
  const [openDialog, setOpenDialog] = useState(false);

  const handleDelete = (index) => {
    setIndex(index);
    setOpenDialog(true);
  };
  const handleEdit = () => {};
  const handleCloseDialog = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        genericPlacementIntentName: props.genericPlacementIntentName,
        appPlacementIntentName: data[index].name,
      };
      apiService
        .deleteAppPlacementIntent(request)
        .then(() => {
          console.log("app placement intent deleted");
          data.splice(index, 1);
          let updatedData = { applications: [...data] };
          setData(updatedData);
        })
        .catch((err) => {
          console.log("Error deleting app placement intent : ", err);
        })
        .finally(() => {
          setIndex(0);
        });
    }
    setOpenDialog(false);
  };
  return (
    <>
      <DeleteDialog
        open={openDialog}
        onClose={handleCloseDialog}
        title={"Delete App Placement Intent"}
        content={`Are you sure you want to delete "${
          data && data[index] ? data[index].name : ""
        }"`}
      />
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <StyledTableCell>Name</StyledTableCell>
              <StyledTableCell>Description</StyledTableCell>
              <StyledTableCell>Intent</StyledTableCell>
              <StyledTableCell>Actions</StyledTableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.map((entry, index) => (
              <StyledTableRow key={entry.name + index}>
                <StyledTableCell>{entry.name}</StyledTableCell>
                <StyledTableCell>{entry.description}</StyledTableCell>
                <StyledTableCell>
                  {entry.allOf &&
                    entry.allOf.map((intent, index) => (
                      <Paper
                        key={index}
                        style={{ width: "max-content" }}
                        variant="outlined"
                      >
                        <label>Cluster Provider :&nbsp;</label>
                        <label style={{ fontWeight: "bold" }}>
                          {intent["provider-name"]}
                        </label>
                        <label>, &nbsp; Cluster :&nbsp;</label>
                        <label style={{ fontWeight: "bold" }}>
                          {intent["cluster-name"]}
                        </label>
                        {intent["cluster-label-name"] && (
                          <>
                            <label>, &nbsp; Labels : </label>
                            <Chip
                              style={{ marginRight: "10px" }}
                              size="small"
                              label={intent["cluster-label-name"]}
                              color="primary"
                              variant="outlined"
                            />
                          </>
                        )}
                      </Paper>
                    ))}
                </StyledTableCell>
                <StyledTableCell>
                  {/* 
                  //edit app placement api has not been implemented yet
                  <IconButton onClick={(e) => handleEdit(index)} title="Edit">
                    <EditIcon color="primary" />
                  </IconButton> */}
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
  );
};

AppPlacementIntentTable.propTypes = {
  data: PropTypes.arrayOf(PropTypes.object).isRequired,
  setData: PropTypes.func.isRequired,
};

export default AppPlacementIntentTable;
