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
import Accordion from "@material-ui/core/Accordion";
import AccordionDetails from "@material-ui/core/AccordionDetails";
import AccordionSummary from "@material-ui/core/AccordionSummary";
import Typography from "@material-ui/core/Typography";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import apiService from "../../services/apiService";
import { Button } from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";
// import EditIcon from "@material-ui/icons/Edit";
import ClusterForm from "./clusters/ClusterForm";
import ClustersTable from "./clusters/ClusterTable";
import DeleteDialog from "../../common/Dialogue";
import Notification from "../../common/Notification";

//import ClusterProviderForm from "../clusterProvider/ClusterProviderForm";

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
export default function ControlledAccordions({ data, setData, ...props }) {
  const classes = useStyles();
  const [expanded, setExpanded] = useState(false);
  const [open, setOpen] = React.useState(false);
  const [formOpen, setFormOpen] = useState(false);
  // const [openProviderForm, setOpenProviderForm] = useState(false);
  const [selectedRowIndex, setSelectedRowIndex] = useState(0);
  const [notificationDetails, setNotificationDetails] = useState({});
  const handleAccordianOpen = (providerRow) => (event, isExpanded) => {
    if (!isExpanded) {
      setExpanded(isExpanded ? providerRow : false);
    } else {
      apiService
        .getClusters(data[providerRow].metadata.name)
        .then((response) => {
          data[providerRow].clusters = response;
          setData([...data]);
        })
        .catch((error) => {
          console.log(error);
        })
        .finally(() => {
          getLabels(providerRow, isExpanded);
          getProviderNetworks(providerRow);
          getNetworks(providerRow);
        });
    }
  };
  const getLabels = (providerRow, isExpanded) => {
    if (data[providerRow].clusters && data[providerRow].clusters.length > 0) {
      data[providerRow].clusters.forEach((cluster) => {
        let request = {
          providerName: data[providerRow].metadata.name,
          clusterName: cluster.metadata.name,
        };
        apiService
          .getClusterLabels(request)
          .then((res) => {
            cluster.labels = res;
          })
          .catch((err) => {
            console.log("error getting cluster label : ", err);
          })
          .finally(() => {
            setData([...data]);
            setExpanded(isExpanded ? providerRow : false);
          });
      });
    } else setExpanded(isExpanded ? providerRow : false);
  };
  const getProviderNetworks = (providerRow) => {
    if (data[providerRow].clusters && data[providerRow].clusters.length > 0) {
      data[providerRow].clusters.forEach((cluster) => {
        let request = {
          providerName: data[providerRow].metadata.name,
          clusterName: cluster.metadata.name,
        };
        apiService
          .getClusterProviderNetworks(request)
          .then((res) => {
            cluster.providerNetworks = res;
          })
          .catch((err) => {
            console.log("error getting cluster provider networks : ", err);
          })
          .finally(() => {
            setData([...data]);
          });
      });
    }
  };
  const getNetworks = (providerRow) => {
    if (data[providerRow].clusters && data[providerRow].clusters.length > 0) {
      data[providerRow].clusters.forEach((cluster) => {
        let request = {
          providerName: data[providerRow].metadata.name,
          clusterName: cluster.metadata.name,
        };
        apiService
          .getClusterNetworks(request)
          .then((res) => {
            cluster.networks = res;
          })
          .catch((err) => {
            console.log("error getting cluster networks : ", err);
          })
          .finally(() => {
            setData([...data]);
          });
      });
    }
  };
  const onAddCluster = (index) => {
    setSelectedRowIndex(index);
    setFormOpen(true);
  };
  const handleDelete = (index) => {
    setSelectedRowIndex(index);
    setOpen(true);
  };
  const handleSubmit = (values, setSubmitting) => {
    let metadata = {};
    if (values.userData) {
      metadata = JSON.parse(values.userData);
    }
    metadata.name = values.name;
    metadata.description = values.description;
    const formData = new FormData();
    formData.append("file", values.file);
    formData.append("metadata", `{"metadata":${JSON.stringify(metadata)}}`);
    formData.append("providerName", data[selectedRowIndex].metadata.name);
    apiService
      .addCluster(formData)
      .then((res) => {
        !data[selectedRowIndex].clusters ||
        data[selectedRowIndex].clusters.length === 0
          ? (data[selectedRowIndex].clusters = [res])
          : data[selectedRowIndex].clusters.push(res);
        setData([...data]);
        setFormOpen(false);
        setNotificationDetails({
          show: true,
          message: `${values.name} cluster added`,
          severity: "success",
        });
      })
      .catch((err) => {
        debugger;
        if (err.response.status === 403) {
          setNotificationDetails({
            show: true,
            message: `${err.response.data}`,
            severity: "error",
          });
          setSubmitting(false);
        }
        console.log("error adding cluster : " + err);
      });
  };
  const handleFormClose = () => {
    setFormOpen(false);
  };
  const handleDeleteCluster = (providerRow, clusterRow) => {
    data[providerRow].clusters.splice(clusterRow, 1);
    setData([...data]);
  };
  const handleUpdateCluster = (providerRow, updatedData) => {
    data[providerRow].clusters = updatedData;
    setData([...data]);
  };
  const handleClose = (el) => {
    if (el.target.innerText === "Delete") {
      apiService
        .deleteClusterProvider(data[selectedRowIndex].metadata.name)
        .then(() => {
          console.log("Cluster Provider deleted");
          data.splice(selectedRowIndex, 1);
          let updatedData = data.slice();
          setData(updatedData);
        })
        .catch((err) => {
          console.log("Error deleting cluster provider : ", err);
        })
        .finally();
    }
    setOpen(false);
    setSelectedRowIndex(0);
  };
  // const handleEdit = (index) => {
  //   setSelectedRowIndex(index);
  //   setOpenProviderForm(true);
  // };
  // const handleCloseProviderForm = () => {
  //   setOpenProviderForm(false);
  // };
  // const handleSubmitProviderForm = (values) => {
  //   let request = {
  //     payload: { metatada: values },
  //     providerName: data[selectedRowIndex].metadata.name,
  //   };
  //   apiService
  //     .updateClusterProvider(request)
  //     .then((res) => {
  //       setData((data) => {
  //         data[selectedRowIndex].metadata = res.metadata;
  //         return data;
  //       });
  //     })
  //     .catch((err) => {
  //       console.log("error updating cluster provider. " + err);
  //     })
  //     .finally(() => {
  //       setOpenProviderForm(false);
  //     });
  // };
  return (
    <>
      <Notification notificationDetails={notificationDetails} />
      {data && data.length > 0 && (
        <div className={classes.root}>
          <ClusterForm
            open={formOpen}
            onClose={handleFormClose}
            onSubmit={handleSubmit}
          />
          {/* <ClusterProviderForm
            open={openProviderForm}
            onClose={handleCloseProviderForm}
            onSubmit={handleSubmitProviderForm}
            item={data[selectedRowIndex]}
          /> */}
          <DeleteDialog
            open={open}
            onClose={handleClose}
            title={"Delete Cluster Provider"}
            content={`Are you sure you want to delete "${
              data[selectedRowIndex] ? data[selectedRowIndex].metadata.name : ""
            }" ?`}
          />
          {data.map((item, index) => (
            <Accordion
              key={item.metadata.name + "" + index}
              expanded={expanded === `${index}`}
              onChange={handleAccordianOpen(`${index}`)}
            >
              <AccordionSummary
                expandIcon={<ExpandMoreIcon />}
                id={`${index}-header`}
              >
                <Typography className={classes.heading}>
                  {item.metadata.name}
                </Typography>
                <Typography className={classes.secondaryHeading}>
                  {item.metadata.description}
                </Typography>
              </AccordionSummary>
              <div style={{ padding: "8px 16px 16px" }}>
                <Button
                  variant="outlined"
                  size="small"
                  color="primary"
                  onClick={() => {
                    onAddCluster(index);
                  }}
                >
                  Add Cluster
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  color="secondary"
                  style={{ float: "right", marginLeft: "10px" }}
                  startIcon={<DeleteIcon />}
                  onClick={() => {
                    handleDelete(index);
                  }}
                >
                  Delete Provider
                </Button>
                {/* 
                //edit cluster provider is not supported by the api yet
                <Button
                  variant="outlined"
                  size="small"
                  color="primary"
                  style={{ float: "right" }}
                  startIcon={<EditIcon />}
                  onClick={() => {
                    handleEdit(index);
                  }}
                >
                  Edit Provider
                </Button> */}
              </div>
              <AccordionDetails>
                {item.clusters && (
                  <ClustersTable
                    clustersData={item.clusters}
                    providerName={item.metadata.name}
                    parentIndex={index}
                    onDeleteCluster={handleDeleteCluster}
                    onUpdateCluster={handleUpdateCluster}
                  />
                )}
                {item.clusters == null && <span>No Clusters</span>}
              </AccordionDetails>
            </Accordion>
          ))}
        </div>
      )}
    </>
  );
}
