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
import PropTypes from "prop-types";
import AddIconOutline from "@material-ui/icons/AddCircleOutline";
import AddIcon from "@material-ui/icons/Add";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import IconButton from "@material-ui/core/IconButton";
// import EditIcon from "@material-ui/icons/Edit";
import Chip from "@material-ui/core/Chip";
import SettingsEthernetIcon from "@material-ui/icons/SettingsEthernet";
import DeleteIcon from "@material-ui/icons/Delete";
import { makeStyles, TextField, Button } from "@material-ui/core";
import NetworkForm from "../networks/NetworkForm";
import apiService from "../../../services/apiService";
import DeleteDialog from "../../../common/Dialogue";
import CancelOutlinedIcon from "@material-ui/icons/CancelOutlined";
import CheckIcon from "@material-ui/icons/CheckCircleOutlineOutlined";
import InfoOutlinedIcon from "@material-ui/icons/InfoOutlined";
import NetworkDetailsDialog from "../../../common/DetailsDialog";
import DoneOutlineIcon from "@material-ui/icons/DoneOutline";
import ClusterForm from "../clusters/ClusterForm";
import Notification from "../../../common/Notification";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
  },
  heading: {
    fontSize: theme.typography.pxToRem(15),
    flexBasis: "33.33%",
    flexShrink: 0,
  },
  secondaryHeading: {
    fontSize: theme.typography.pxToRem(15),
    color: theme.palette.text.secondary,
  },
}));

const ClusterTable = ({ clustersData, ...props }) => {
  const classes = useStyles();
  const [formOpen, setformOpen] = useState(false);
  const [networkDetailsOpen, setNetworkDetailsOpen] = useState(false);
  const [network, setNetwork] = useState({});
  const [activeRowIndex, setActiveRowIndex] = useState(0);
  const [activeNetwork, setActiveNetwork] = useState({});
  const [open, setOpen] = useState(false);
  const [openDeleteNetwork, setOpenDeleteNetwork] = useState(false);
  const [showAddLabel, setShowAddLabel] = useState(false);
  const [labelInput, setLabelInput] = useState("");
  //   const [clusterFormOpen, setClusterFormOpen] = useState(false);
  const [notificationDetails, setNotificationDetails] = useState({});
  const handleFormClose = () => {
    setformOpen(false);
  };
  const handleSubmit = (data) => {
    let networkSpec = JSON.parse(data.spec);
    let payload = {
      metadata: { name: data.name, description: data.description },
      spec: networkSpec,
    };
    let request = {
      providerName: props.providerName,
      clusterName: clustersData[activeRowIndex].metadata.name,
      networkType: data.type,
      payload: payload,
    };
    apiService
      .addNetwork(request)
      .then((res) => {
        let networkType =
          data.type === "networks" ? "networks" : "providerNetworks";
        !clustersData[activeRowIndex][networkType] ||
        clustersData[activeRowIndex][networkType] === null
          ? (clustersData[activeRowIndex][networkType] = [res])
          : clustersData[activeRowIndex][networkType].push(res);
      })
      .catch((err) => {
        console.log("error adding cluster network : ", err);
      })
      .finally(() => {
        setActiveRowIndex(0);
        setformOpen(false);
      });
  };
  const handleAddNetwork = (index) => {
    setActiveRowIndex(index);
    setformOpen(true);
  };
  const handleDeleteLabel = (index, label, labelIndex) => {
    let request = {
      providerName: props.providerName,
      clusterName: clustersData[index].metadata.name,
      labelName: label,
    };
    apiService
      .deleteClusterLabel(request)
      .then((res) => {
        console.log("label deleted");
        clustersData[index].labels.splice(labelIndex, 1);
        props.onUpdateCluster(props.parentIndex, clustersData);
      })
      .catch((err) => {
        console.log("error deleting label : ", err);
      });
  };
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        providerName: props.providerName,
        clusterName: clustersData[activeRowIndex].metadata.name,
      };
      apiService
        .deleteCluster(request)
        .then(() => {
          console.log("cluster deleted");
          props.onDeleteCluster(props.parentIndex, activeRowIndex);
        })
        .catch((err) => {
          console.log("Error deleting cluster : ", +err);
          setNotificationDetails({
            show: true,
            message: "Unable to remove cluster",
            severity: "error",
          });
        });
    }
    setOpen(false);
    setActiveRowIndex(0);
  };

  const handleCloseDeleteNetwork = (el) => {
    if (el.target.innerText === "Delete") {
      let networkName =
        clustersData[activeRowIndex][activeNetwork.networkType][
          activeNetwork.networkIndex
        ].metadata.name;
      let networkType =
        activeNetwork.networkType === "providerNetworks"
          ? "provider-networks"
          : "networks";
      let request = {
        providerName: props.providerName,
        clusterName: clustersData[activeRowIndex].metadata.name,
        networkType: networkType,
        networkName: networkName,
      };
      apiService
        .deleteClusterNetwork(request)
        .then(() => {
          console.log("cluster network deleted");
          clustersData[activeRowIndex][activeNetwork.networkType].splice(
            activeNetwork.networkIndex,
            1
          );
        })
        .catch((err) => {
          console.log("Error deleting cluster network : ", err);
        })
        .finally(() => {
          setActiveRowIndex(0);
          setActiveNetwork({});
        });
    }
    setOpenDeleteNetwork(false);
  };
  const handleDeleteCluster = (index) => {
    setActiveRowIndex(index);
    setOpen(true);
  };
  const handleAddLabel = (index) => {
    if (labelInput !== "") {
      let request = {
        providerName: props.providerName,
        clusterName: clustersData[activeRowIndex].metadata.name,
        payload: { "label-name": labelInput },
      };
      apiService
        .addClusterLabel(request)
        .then((res) => {
          !clustersData[index].labels || clustersData[index].labels === null
            ? (clustersData[index].labels = [res])
            : clustersData[index].labels.push(res);
        })
        .catch((err) => {
          console.log("error adding label", err);
        })
        .finally(() => {
          setShowAddLabel(!showAddLabel);
        });
    }
  };

  const handleToggleAddLabel = (index) => {
    setShowAddLabel(showAddLabel === index ? false : index);
    setActiveRowIndex(index);
    setLabelInput("");
  };
  const handleLabelInputChange = (event) => {
    setLabelInput(event.target.value);
  };

  const handleNetworkDetailOpen = (network) => {
    setNetwork(network);
    setNetworkDetailsOpen(true);
  };
  const handleDeleteNetwork = (
    index,
    networkIndex,
    networkType,
    networkName
  ) => {
    setActiveNetwork({
      networkIndex: networkIndex,
      networkType: networkType,
      name: networkName,
    });
    setActiveRowIndex(index);
    setOpenDeleteNetwork(true);
  };
  const applyNetworkConfig = (clusterName) => {
    let request = {
      providerName: props.providerName,
      clusterName: clusterName,
    };
    apiService
      .applyNetworkConfig(request)
      .then((res) => {
        setNotificationDetails({
          show: true,
          message: "Network configuration applied",
          severity: "success",
        });
        console.log("Network config applied");
      })
      .catch((err) => {
        setNotificationDetails({
          show: true,
          message: "Error applying network configuration",
          severity: "error",
        });
        console.log("Error applying network config : ", err);
        if (err.response)
          console.log("Network config applied" + err.response.data);
        else console.log("Network config applied" + err);
      });
  };
  //   const handleClusterFormClose = () => {
  //     setClusterFormOpen(false);
  //   };
  //   const handleClusterSubmit = (values) => {
  //     const formData = new FormData();
  //     if (values.file) formData.append("file", values.file);
  //     formData.append(
  //       "metadata",
  //       `{"metadata":{ "name": "${values.name}", "description": "${values.description}" }}`
  //     );
  //     formData.append("providerName", props.providerName);
  //     apiService
  //       .updateCluster(formData)
  //       .then((res) => {
  //         clustersData[activeRowIndex].metadata = res.metadata;
  //         props.onUpdateCluster(props.parentIndex, clustersData);
  //       })
  //       .catch((err) => {
  //         console.log("error updating cluster : ", err);
  //       })
  //       .finally(() => {
  //         handleClusterFormClose();
  //       });
  //   };
  //disabling as edit is not supported yet by the api yet
  //   const handleEditCluster = (index) => {
  //     setActiveRowIndex(index);
  //     setClusterFormOpen(true);
  //   };
  return (
    <>
      <Notification notificationDetails={notificationDetails} />
      {clustersData && clustersData.length > 0 && (
        <>
          {/* <ClusterForm
            item={clustersData[activeRowIndex]}
            open={clusterFormOpen}
            onClose={handleClusterFormClose}
            onSubmit={handleClusterSubmit}
          /> */}
          <NetworkDetailsDialog
            onClose={setNetworkDetailsOpen}
            open={networkDetailsOpen}
            item={network}
            type="Network"
          />
          <NetworkForm
            onClose={handleFormClose}
            onSubmit={handleSubmit}
            open={formOpen}
          />
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Delete Cluster"}
            content={`Are you sure you want to delete "${
              clustersData[activeRowIndex]
                ? clustersData[activeRowIndex].metadata.name
                : ""
            }" ?`}
          />
          <DeleteDialog
            open={openDeleteNetwork}
            onClose={handleCloseDeleteNetwork}
            title={"Delete Network"}
            content={`Are you sure you want to delete "${activeNetwork.name}" ?`}
          />
          <TableContainer>
            <Table className={classes.table}>
              <TableHead>
                <TableRow>
                  <TableCell style={{ width: "10%" }}>Name</TableCell>
                  <TableCell style={{ width: "15%" }}>Description</TableCell>
                  <TableCell style={{ width: "20%" }}>Networks </TableCell>
                  <TableCell style={{ width: "35%" }}>Labels </TableCell>
                  <TableCell style={{ width: "20%" }}>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {clustersData.map((row, index) => (
                  <TableRow key={row.metadata.name + "" + index}>
                    <TableCell>{row.metadata.name}</TableCell>
                    <TableCell>{row.metadata.description}</TableCell>
                    <TableCell>
                      <div>
                        {row.providerNetworks &&
                          row.providerNetworks.length > 0 &&
                          row.providerNetworks.map(
                            (providerNetwork, providerNetworkIndex) => (
                              <Chip
                                key={
                                  providerNetwork.metadata.name +
                                  "" +
                                  providerNetworkIndex
                                }
                                size="small"
                                icon={
                                  <InfoOutlinedIcon
                                    onClick={() => {
                                      handleNetworkDetailOpen(providerNetwork);
                                    }}
                                    style={{ cursor: "pointer" }}
                                  />
                                }
                                onDelete={(e) => {
                                  handleDeleteNetwork(
                                    index,
                                    providerNetworkIndex,
                                    "providerNetworks",
                                    providerNetwork.metadata.name
                                  );
                                }}
                                label={providerNetwork.metadata.name}
                                style={{
                                  marginRight: "10px",
                                  marginBottom: "5px",
                                }}
                              />
                            )
                          )}

                        {row.networks &&
                          row.networks.length > 0 &&
                          row.networks.map((network, networkIndex) => (
                            <Chip
                              key={network.metadata.name + "" + networkIndex}
                              size="small"
                              icon={
                                <InfoOutlinedIcon
                                  onClick={() => {
                                    handleNetworkDetailOpen(network);
                                  }}
                                  style={{ cursor: "pointer" }}
                                />
                              }
                              onDelete={(e) => {
                                handleDeleteNetwork(
                                  index,
                                  networkIndex,
                                  "networks",
                                  network.metadata.name
                                );
                              }}
                              label={network.metadata.name}
                              style={{
                                marginRight: "10px",
                                marginBottom: "5px",
                              }}
                              color="secondary"
                            />
                          ))}
                      </div>
                    </TableCell>
                    <TableCell>
                      {row.labels &&
                        row.labels.length > 0 &&
                        row.labels.map((label, labelIndex) => (
                          <Chip
                            key={label["label-name"] + "" + labelIndex}
                            size="small"
                            icon={<SettingsEthernetIcon />}
                            label={label["label-name"]}
                            onDelete={(e) => {
                              handleDeleteLabel(
                                index,
                                label["label-name"],
                                labelIndex
                              );
                            }}
                            color="primary"
                            style={{ marginRight: "10px" }}
                          />
                        ))}
                      {showAddLabel === index && (
                        <TextField
                          style={{ height: "24px" }}
                          size="small"
                          value={labelInput}
                          onChange={handleLabelInputChange}
                          id="outlined-basic"
                          label="Add label"
                          variant="outlined"
                        />
                      )}
                      {showAddLabel === index && (
                        <IconButton
                          color="primary"
                          onClick={() => {
                            handleAddLabel(index);
                          }}
                        >
                          <CheckIcon />
                        </IconButton>
                      )}
                      <IconButton
                        color="primary"
                        onClick={() => {
                          handleToggleAddLabel(index);
                        }}
                      >
                        {!(showAddLabel === index) && <AddIconOutline />}
                        {showAddLabel === index && (
                          <CancelOutlinedIcon color="secondary" />
                        )}
                      </IconButton>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="outlined"
                        startIcon={<AddIcon />}
                        size="small"
                        color="primary"
                        title="Add Network"
                        onClick={() => {
                          handleAddNetwork(index);
                        }}
                      >
                        Network
                      </Button>
                      <IconButton
                        color="primary"
                        disabled={
                          !(
                            (row.networks && row.networks.length > 0) ||
                            (row.providerNetworks &&
                              row.providerNetworks.length > 0)
                          )
                        }
                        onClick={() => {
                          applyNetworkConfig(row.metadata.name);
                        }}
                        title="Apply Network Configuration"
                      >
                        <DoneOutlineIcon />
                      </IconButton>
                      {/* 
                      //disabling as edit is not supported yet by the api yet
                        <IconButton
                            title="Edit"
                            onClick={() => { handleEditCluster(index) }}
                            color="primary">
                            <EditIcon />
                        </IconButton> */}
                      <IconButton
                        title="Delete"
                        color="secondary"
                        disabled={
                          (row.networks && row.networks.length > 0) ||
                          (row.providerNetworks &&
                            row.providerNetworks.length > 0) ||
                          (row.labels && row.labels.length > 0)
                        }
                        onClick={() => {
                          handleDeleteCluster(index);
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </>
      )}
      {(!clustersData || clustersData.length === 0) && <span>No Clusters</span>}
    </>
  );
};
ClusterTable.propTypes = {
  clusters: PropTypes.arrayOf(PropTypes.object),
};
export default ClusterTable;
