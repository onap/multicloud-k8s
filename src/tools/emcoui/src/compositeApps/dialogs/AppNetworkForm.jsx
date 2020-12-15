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
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import { Grid, IconButton } from "@material-ui/core";
import { TextField, Select, MenuItem, InputLabel } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import CardContent from "@material-ui/core/CardContent";
import Card from "@material-ui/core/Card";
import apiService from "../../services/apiService";
import DeleteIcon from "@material-ui/icons/Delete";
import { Formik } from "formik";
import Notification from "../../common/Notification";

function NetworkForm({ formikProps, ...props }) {
  const [clusters, setClusters] = useState(props.clusters);
  const [notificationDetails, setNotificationDetails] = useState({});
  const useStyles = makeStyles({
    root: {
      minWidth: 275,
    },
    title: {
      fontSize: 14,
    },
    pos: {
      marginBottom: 12,
    },
  });

  const handleAddNetworkInterface = (providerIndex, clusterIndex, values) => {
    let updatedFields = [];
    if (
      values.apps[props.index].clusters[providerIndex].selectedClusters[
        clusterIndex
      ].interfaces
    ) {
      updatedFields = [
        ...values.apps[props.index].clusters[providerIndex].selectedClusters[
          clusterIndex
        ].interfaces,
        {
          networkName: "",
          ip: "",
          subnet: "",
        },
      ];
    } else {
      updatedFields = [
        {
          networkName: "",
          ip: "",
          subnet: "",
        },
      ];
    }

    let request = {
      providerName: values.apps[props.index].clusters[providerIndex].provider,
      clusterName:
        values.apps[props.index].clusters[providerIndex].selectedClusters[
          clusterIndex
        ].name,
    };
    apiService
      .getClusterProviderNetworks(request)
      .then((networks) => {
        let networkData = [];
        if (networks && networks.length > 0) {
          networks.forEach((network) => {
            networkData.push({
              name: network.metadata.name,
              subnets: network.spec.ipv4Subnets,
            });
          });
        }

        apiService
          .getClusterNetworks(request)
          .then((clusterNetworks) => {
            if (clusterNetworks && clusterNetworks.length > 0) {
              clusterNetworks.forEach((clusterNetwork) => {
                networkData.push({
                  name: clusterNetwork.metadata.name,
                  subnets: clusterNetwork.spec.ipv4Subnets,
                });
              });
            }
            //add interface entry onyl of there is atlease one available network
            if (networkData.length > 0) {
              setClusters((clusters) => {
                clusters[providerIndex].selectedClusters[
                  clusterIndex
                ].interfaces = updatedFields;
                clusters[providerIndex].selectedClusters[
                  clusterIndex
                ].networks = networkData;
                clusters[providerIndex].selectedClusters[
                  clusterIndex
                ].availableNetworks = getAvailableNetworks(
                  clusters[providerIndex].selectedClusters[clusterIndex]
                );
                return clusters;
              });
              formikProps.setFieldValue(
                `apps[${props.index}].clusters[${providerIndex}].selectedClusters[${clusterIndex}].interfaces`,
                updatedFields
              );
            } else {
              setNotificationDetails({
                show: true,
                message: `No network available for this cluster`,
                severity: "warning",
              });
            }
          })
          .catch((err) => {
            console.log("error getting cluster networks : ", err);
          });
      })
      .catch((err) => {
        console.log("error getting cluster provider networks : ", err);
      })
      .finally(() => {
        return updatedFields;
      });
  };

  const handleSelectNetowrk = (
    e,
    providerIndex,
    clusterIndex,
    interfaceIndex
  ) => {
    setClusters((clusters) => {
      clusters[providerIndex].selectedClusters[clusterIndex].interfaces[
        interfaceIndex
      ] = {
        networkName: e.target.value,
        ip: "",
        subnet: "",
      };
      clusters[providerIndex].selectedClusters[
        clusterIndex
      ].availableNetworks = getAvailableNetworks(
        clusters[providerIndex].selectedClusters[clusterIndex],
        "handleAddNetworkInterface"
      );
      return clusters;
    });
    formikProps.handleChange(e);
  };
  const handleRemoveNetwork = (providerIndex, clusterIndex, interfaceIndex) => {
    setClusters((clusters) => {
      clusters[providerIndex].selectedClusters[clusterIndex].interfaces.splice(
        interfaceIndex,
        1
      );
      clusters[providerIndex].selectedClusters[
        clusterIndex
      ].availableNetworks = getAvailableNetworks(
        clusters[providerIndex].selectedClusters[clusterIndex]
      );
      return clusters;
    });
    formikProps.setFieldValue(
      `apps[${props.index}].clusters[${providerIndex}].selectedClusters[${clusterIndex}].interfaces`,
      clusters[providerIndex].selectedClusters[clusterIndex].interfaces
    );
  };
  const getAvailableNetworks = (cluster) => {
    let availableNetworks = [];
    cluster.networks.forEach((network) => {
      let match = false;
      cluster.interfaces.forEach((networkInterface) => {
        if (network.name === networkInterface.networkName) {
          match = true;
          return;
        }
      });
      if (!match) availableNetworks.push(network);
    });
    return availableNetworks;
  };

  const classes = useStyles();
  return (
    <>
      <Notification notificationDetails={notificationDetails} />
      <Grid
        key="networkForm"
        container
        spacing={3}
        style={{
          height: "400px",
          overflowY: "auto",
          width: "100%",
          scrollbarWidth: "thin",
        }}
      >
        {(!clusters || clusters.length < 1) && (
          <Grid item xs={12}>
            <Typography variant="h6">No clusters selected</Typography>
          </Grid>
        )}
        {clusters &&
          clusters.map((cluster, providerIndex) => (
            <Grid key={cluster.provider + providerIndex} item xs={12}>
              <Card className={classes.root}>
                <CardContent>
                  <Grid container spacing={2}>
                    <Grid item xs={12}>
                      <Typography
                        className={classes.title}
                        color="textSecondary"
                        gutterBottom
                      >
                        {cluster.provider}
                      </Typography>
                    </Grid>
                    {cluster.selectedClusters.map(
                      (selectedCluster, clusterIndex) => (
                        <React.Fragment key={selectedCluster.name}>
                          <Grid item xs={12}>
                            <Typography>{selectedCluster.name}</Typography>
                          </Grid>
                          <Formik>
                            {() => {
                              const {
                                values,
                                errors,
                                handleChange,
                                handleBlur,
                              } = formikProps;
                              return (
                                <>
                                  {selectedCluster.interfaces &&
                                  selectedCluster.interfaces.length > 0
                                    ? selectedCluster.interfaces.map(
                                        (networkInterface, interfaceIndex) => (
                                          <Grid
                                            spacing={1}
                                            container
                                            item
                                            key={interfaceIndex}
                                            xs={12}
                                          >
                                            <Grid item xs={4}>
                                              <InputLabel id="network-select-label">
                                                Network
                                              </InputLabel>
                                              <Select
                                                fullWidth
                                                labelId="network-select-label"
                                                id="network-select"
                                                name={`apps[${props.index}].clusters[${providerIndex}].selectedClusters[${clusterIndex}].interfaces[${interfaceIndex}].networkName`}
                                                value={
                                                  values.apps[props.index]
                                                    .clusters[providerIndex]
                                                    .selectedClusters[
                                                    clusterIndex
                                                  ].interfaces[interfaceIndex]
                                                    .networkName
                                                }
                                                onChange={(e) => {
                                                  handleSelectNetowrk(
                                                    e,
                                                    providerIndex,
                                                    clusterIndex,
                                                    interfaceIndex
                                                  );
                                                }}
                                              >
                                                {values.apps[props.index]
                                                  .clusters[providerIndex]
                                                  .selectedClusters[
                                                  clusterIndex
                                                ].interfaces[interfaceIndex]
                                                  .networkName && (
                                                  <MenuItem
                                                    key={
                                                      values.apps[props.index]
                                                        .clusters[providerIndex]
                                                        .selectedClusters[
                                                        clusterIndex
                                                      ].interfaces[
                                                        interfaceIndex
                                                      ].networkName
                                                    }
                                                    value={
                                                      values.apps[props.index]
                                                        .clusters[providerIndex]
                                                        .selectedClusters[
                                                        clusterIndex
                                                      ].interfaces[
                                                        interfaceIndex
                                                      ].networkName
                                                    }
                                                  >
                                                    {
                                                      values.apps[props.index]
                                                        .clusters[providerIndex]
                                                        .selectedClusters[
                                                        clusterIndex
                                                      ].interfaces[
                                                        interfaceIndex
                                                      ].networkName
                                                    }
                                                  </MenuItem>
                                                )}
                                                {selectedCluster.availableNetworks &&
                                                  selectedCluster.availableNetworks.map(
                                                    (network) => (
                                                      <MenuItem
                                                        key={network.name}
                                                        value={network.name}
                                                      >
                                                        {network.name}
                                                      </MenuItem>
                                                    )
                                                  )}
                                              </Select>
                                            </Grid>

                                            <Grid item xs={4}>
                                              <InputLabel id="subnet-select-label">
                                                Subnet
                                              </InputLabel>
                                              <Select
                                                fullWidth
                                                labelId="subnet-select-label"
                                                id="subnet-select-label"
                                                name={`apps[${props.index}].clusters[${providerIndex}].selectedClusters[${clusterIndex}].interfaces[${interfaceIndex}].subnet`}
                                                value={
                                                  values.apps[props.index]
                                                    .clusters[providerIndex]
                                                    .selectedClusters[
                                                    clusterIndex
                                                  ].interfaces[interfaceIndex]
                                                    .subnet
                                                }
                                                onChange={handleChange}
                                              >
                                                {values.apps[props.index]
                                                  .clusters[providerIndex]
                                                  .selectedClusters[
                                                  clusterIndex
                                                ].interfaces[interfaceIndex]
                                                  .networkName === ""
                                                  ? null
                                                  : selectedCluster.networks
                                                      .filter(
                                                        (network) =>
                                                          network.name ===
                                                          values.apps[
                                                            props.index
                                                          ].clusters[
                                                            providerIndex
                                                          ].selectedClusters[
                                                            clusterIndex
                                                          ].interfaces[
                                                            interfaceIndex
                                                          ].networkName
                                                      )[0]
                                                      .subnets.map((subnet) => (
                                                        <MenuItem
                                                          key={subnet.name}
                                                          value={subnet.name}
                                                        >
                                                          {subnet.name}(
                                                          {subnet.subnet})
                                                        </MenuItem>
                                                      ))}
                                              </Select>
                                            </Grid>
                                            <Grid item xs={3}>
                                              <TextField
                                                width={"65%"}
                                                name={`apps[${props.index}].clusters[${providerIndex}].selectedClusters[${clusterIndex}].interfaces[${interfaceIndex}].ip`}
                                                onBlur={handleBlur}
                                                id="ip"
                                                label="IP Address"
                                                value={
                                                  values.apps[props.index]
                                                    .clusters[providerIndex]
                                                    .selectedClusters[
                                                    clusterIndex
                                                  ].interfaces[interfaceIndex]
                                                    .ip
                                                }
                                                onChange={handleChange}
                                                helperText={
                                                  (errors.apps &&
                                                    errors.apps[props.index] &&
                                                    errors.apps[props.index]
                                                      .clusters &&
                                                    errors.apps[props.index]
                                                      .clusters[clusterIndex] &&
                                                    errors.apps[props.index]
                                                      .clusters[clusterIndex]
                                                      .selectedClusters[
                                                      clusterIndex
                                                    ] &&
                                                    errors.apps[props.index]
                                                      .clusters[clusterIndex]
                                                      .selectedClusters[
                                                      clusterIndex
                                                    ].interfaces[
                                                      interfaceIndex
                                                    ] &&
                                                    errors.apps[props.index]
                                                      .clusters[clusterIndex]
                                                      .selectedClusters[
                                                      clusterIndex
                                                    ].interfaces[interfaceIndex]
                                                      .ip) ||
                                                  "blank for auto assign"
                                                }
                                                error={
                                                  errors.apps &&
                                                  errors.apps[props.index] &&
                                                  errors.apps[props.index]
                                                    .clusters &&
                                                  errors.apps[props.index]
                                                    .clusters[clusterIndex] &&
                                                  errors.apps[props.index]
                                                    .clusters[clusterIndex]
                                                    .selectedClusters[
                                                    clusterIndex
                                                  ] &&
                                                  errors.apps[props.index]
                                                    .clusters[clusterIndex]
                                                    .selectedClusters[
                                                    clusterIndex
                                                  ].interfaces[
                                                    interfaceIndex
                                                  ] &&
                                                  errors.apps[props.index]
                                                    .clusters[clusterIndex]
                                                    .selectedClusters[
                                                    clusterIndex
                                                  ].interfaces[interfaceIndex]
                                                    .ip &&
                                                  true
                                                }
                                              />
                                            </Grid>
                                            <Grid item xs={1}>
                                              <IconButton
                                                color="secondary"
                                                onClick={() => {
                                                  handleRemoveNetwork(
                                                    providerIndex,
                                                    clusterIndex,
                                                    interfaceIndex
                                                  );
                                                }}
                                              >
                                                <DeleteIcon fontSize="small" />
                                              </IconButton>
                                            </Grid>
                                          </Grid>
                                        )
                                      )
                                    : null}
                                  <Grid
                                    key={selectedCluster.name + "addButton"}
                                    item
                                    xs={12}
                                  >
                                    <Button
                                      variant="outlined"
                                      size="small"
                                      fullWidth
                                      color="primary"
                                      disabled={
                                        selectedCluster.interfaces &&
                                        selectedCluster.interfaces.length > 0 &&
                                        selectedCluster.networks.length ===
                                          selectedCluster.interfaces.length
                                      }
                                      onClick={() => {
                                        handleAddNetworkInterface(
                                          providerIndex,
                                          clusterIndex,
                                          values
                                        );
                                      }}
                                      startIcon={<AddIcon />}
                                    >
                                      Add Network Interface
                                    </Button>
                                  </Grid>
                                </>
                              );
                            }}
                          </Formik>
                        </React.Fragment>
                      )
                    )}
                  </Grid>
                </CardContent>
              </Card>
            </Grid>
          ))}
      </Grid>
    </>
  );
}

export default NetworkForm;
