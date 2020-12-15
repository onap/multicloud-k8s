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
import React, { useEffect, useState } from "react";
import { Formik } from "formik";
import * as Yup from "yup";
import AppForm from "./DigFormApp";
import apiService from "../services/apiService";

import { Button, DialogActions, Grid } from "@material-ui/core";

DigFormIntents.propTypes = {};
const schema = Yup.object({
  apps: Yup.array()
    .of(
      Yup.object({
        clusters: Yup.array()
          .of(
            Yup.object({
              provider: Yup.string(),
              selectedClusters: Yup.array().of(
                Yup.object({
                  name: Yup.string(),
                  interfaces: Yup.array().of(
                    Yup.object({
                      networkName: Yup.string().required(),
                      subnet: Yup.string().required(),
                      ip: Yup.string().matches(
                        /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
                        "invalid ip address"
                      ),
                    })
                  ),
                })
              ),
            })
          )
          .required("Select at least one cluster"),
      })
    )
    .required("At least one app is required"),
});

function DigFormIntents(props) {
  const { onSubmit, appsData } = props;
  const [isLoading, setIsloading] = useState(true);
  const [clusterProviders, setClusterProviders] = useState([]);
  let initialValues = { apps: appsData };
  useEffect(() => {
    let clusterProviderData = [];
    apiService
      .getClusterProviders()
      .then((res) => {
        res.forEach((clusterProvider, providerIndex) => {
          clusterProviderData.push({
            name: clusterProvider.metadata.name,
            clusters: [],
          });
          apiService
            .getClusters(clusterProvider.metadata.name)
            .then((clusters) => {
              clusters.forEach((cluster) => {
                clusterProviderData[providerIndex].clusters.push({
                  name: cluster.metadata.name,
                  description: cluster.metadata.description,
                });
              });
              if (providerIndex + 1 === res.length) {
                setClusterProviders(clusterProviderData);
                setIsloading(false);
              }
            })
            .catch((err) => {
              console.log(
                `error getting clusters for ${clusterProvider.metadata.name} : ` +
                  err
              );
            });
        });
      })
      .catch((err) => {
        console.log("error getting cluster providers : " + err);
      });
  }, []);
  useEffect(() => {}, []);

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={(values) => {
        values.compositeAppVersion = onSubmit(values);
      }}
      validationSchema={schema}
    >
      {(formikProps) => {
        const {
          values,
          isSubmitting,
          handleChange,
          handleSubmit,
        } = formikProps;
        return (
          !isLoading && (
            <form noValidate onSubmit={handleSubmit} onChange={handleChange}>
              <Grid container spacing={4} justify="center">
                {initialValues.apps &&
                  initialValues.apps.length > 0 &&
                  initialValues.apps.map((app, index) => (
                    <Grid key={index} item sm={12} xs={12}>
                      <AppForm
                        clusterProviders={clusterProviders}
                        formikProps={formikProps}
                        name={app.metadata.name}
                        description={app.metadata.description}
                        index={index}
                        initialValues={values}
                      />
                    </Grid>
                  ))}

                <Grid item xs={12}>
                  <DialogActions>
                    <Button
                      autoFocus
                      onClick={props.onClickBack}
                      color="secondary"
                    >
                      Back
                    </Button>
                    <Button
                      autoFocus
                      type="submit"
                      color="primary"
                      disabled={isSubmitting}
                    >
                      Submit
                    </Button>
                  </DialogActions>
                </Grid>
              </Grid>
            </form>
          )
        );
      }}
    </Formik>
  );
}

export default DigFormIntents;
