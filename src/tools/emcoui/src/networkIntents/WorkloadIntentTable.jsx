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
  Chip,
  TableCell,
} from "@material-ui/core";
import Paper from "@material-ui/core/Paper";
import TableBody from "@material-ui/core/TableBody";
import EditIcon from "@material-ui/icons/Edit";
import DeleteIcon from "@material-ui/icons/Delete";
import PropTypes from "prop-types";
import apiService from "../services/apiService";
import DeleteDialog from "../common/Dialogue";
import InfoOutlinedIcon from "@material-ui/icons/InfoOutlined";
import IconButton from "@material-ui/core/IconButton";
import AddIconOutline from "@material-ui/icons/AddCircleOutline";
import Form from "./InterfaceForm";
import InterfaceDetailsDialog from "../common/DetailsDialog";

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

const WokloadIntentTable = ({ data, setData, ...props }) => {
  const [formOpen, setFormOpen] = useState(false);
  const [index, setIndex] = useState(0);
  const [openDialog, setOpenDialog] = useState(false);
  const [openInterfaceDetails, setOpenInterfaceDetails] = useState(false);
  const [selectedInterface, setSelectedInterface] = useState({});
  const [openInterfaceDialog, setOpenInterfaceDialog] = useState(false);
  const handleDelete = (index) => {
    setIndex(index);
    setOpenDialog(true);
  };
  const handleEdit = () => {};
  const handleInterfaceDetailOpen = (entry) => {
    setSelectedInterface(entry);
    setOpenInterfaceDetails(true);
  };

  const handleDeleteInterface = (index, entry) => {
    setIndex(index);
    setSelectedInterface(entry);
    setOpenInterfaceDialog(true);
  };
  const handleAddInterface = (index) => {
    setIndex(index);
    setFormOpen(true);
  };
  const handleCloseForm = () => {
    setFormOpen(false);
  };
  const handleCloseInterfaceDialog = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        networkControllerIntentName: props.networkControllerIntentName,
        workloadIntentName: data[index].metadata.name,
        interfaceName: selectedInterface.metadata.name,
      };
      apiService
        .deleteInterface(request)
        .then(() => {
          console.log("Interface deleted");
          let updatedInterfaceData = data[index].interfaces.filter(function (
            obj
          ) {
            return obj.metadata.name !== selectedInterface.metadata.name;
          });
          data[index].interfaces = updatedInterfaceData;
          setData([...data]);
        })
        .catch((err) => {
          console.log("Error deleting interface : ", err);
        })
        .finally(() => {
          setIndex(0);
          setSelectedInterface({});
        });
    }
    setOpenInterfaceDialog(false);
  };
  const handleCloseDialog = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        networkControllerIntentName: props.networkControllerIntentName,
        workloadIntentName: data[index].metadata.name,
      };
      apiService
        .deleteWorkloadIntent(request)
        .then(() => {
          console.log("workload intent deleted");
          data.splice(index, 1);
          setData([...data]);
        })
        .catch((err) => {
          console.log("Error deleting workload intent : ", err);
        })
        .finally(() => {
          setIndex(0);
        });
    }
    setOpenDialog(false);
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
      networkControllerIntentName: props.networkControllerIntentName,
      workloadIntentName: data[index].metadata.name,
    };
    apiService
      .addInterface(request)
      .then((res) => {
        if (data[index].interfaces && data[index].interfaces.length > 0) {
          data[index].interfaces.push(res);
        } else {
          data[index].interfaces = [res];
        }
        setData([...data]);
      })
      .catch((err) => {
        console.log("error creating composite profile : ", err);
      })
      .finally(() => {
        setFormOpen(false);
      });
  };
  return (
    <>
      <InterfaceDetailsDialog
        open={openInterfaceDetails}
        onClose={setOpenInterfaceDetails}
        item={selectedInterface}
        type="Interface"
      />
      <Form open={formOpen} onClose={handleCloseForm} onSubmit={handleSubmit} />
      <DeleteDialog
        open={openDialog}
        onClose={handleCloseDialog}
        title={"Delete Profile"}
        content={`Are you sure you want to delete "${
          data && data[index] ? data[index].metadata.name : ""
        }"`}
      />
      <DeleteDialog
        open={openInterfaceDialog}
        onClose={handleCloseInterfaceDialog}
        title={"Delete Interface"}
        content={`Are you sure you want to delete "${
          selectedInterface.metadata ? selectedInterface.metadata.name : ""
        }"`}
      />
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <StyledTableCell>Name</StyledTableCell>
              <StyledTableCell>Description</StyledTableCell>
              <StyledTableCell>App</StyledTableCell>
              <StyledTableCell>Workload Resource</StyledTableCell>
              <StyledTableCell style={{ width: "27%" }}>
                Interfaces
              </StyledTableCell>
              <StyledTableCell>Actions</StyledTableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.map((entry, index) => (
              <StyledTableRow key={entry.metadata.name + index}>
                <StyledTableCell>{entry.metadata.name}</StyledTableCell>
                <StyledTableCell>{entry.metadata.description}</StyledTableCell>
                <StyledTableCell>
                  {entry.spec["application-name"]}
                </StyledTableCell>
                <StyledTableCell>
                  {entry.spec["workload-resource"]}
                </StyledTableCell>
                <StyledTableCell>
                  {entry.interfaces &&
                    entry.interfaces.length > 0 &&
                    entry.interfaces.map((interfaceEntry, interfacekIndex) => (
                      <Chip
                        key={
                          interfaceEntry.metadata.name + "" + interfacekIndex
                        }
                        size="small"
                        icon={
                          <InfoOutlinedIcon
                            onClick={() => {
                              handleInterfaceDetailOpen(interfaceEntry);
                            }}
                            style={{ cursor: "pointer" }}
                          />
                        }
                        onDelete={(e) => {
                          handleDeleteInterface(index, interfaceEntry);
                        }}
                        label={interfaceEntry.spec.ipAddress}
                        style={{ marginRight: "10px", marginBottom: "5px" }}
                      />
                    ))}
                  <IconButton
                    color="primary"
                    onClick={() => {
                      handleAddInterface(index);
                    }}
                  >
                    <AddIconOutline />
                  </IconButton>
                </StyledTableCell>
                <StyledTableCell>
                  {/* 
                  //edit workload intent api has not been added yet
                  <IconButton onClick={(e) => handleEdit(index)} title="Edit" >
                      <EditIcon color="primary" />
                  </IconButton> */}
                  <IconButton
                    color="secondary"
                    disabled={entry.interfaces && entry.interfaces.length > 0}
                    onClick={(e) => handleDelete(index)}
                    title="Delete"
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
  );
};
WokloadIntentTable.propTypes = {
  data: PropTypes.arrayOf(PropTypes.object).isRequired,
  setData: PropTypes.func.isRequired,
};
export default WokloadIntentTable;
