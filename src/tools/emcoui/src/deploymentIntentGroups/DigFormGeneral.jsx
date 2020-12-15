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

import {
  Button,
  DialogActions,
  FormControl,
  FormHelperText,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from "@material-ui/core";

const schema = Yup.object({
  name: Yup.string().required(),
  description: Yup.string(),
  version: Yup.string()
    .matches(/^[A-Za-z0-9\\s]+$/, "Special characters and space not allowed")
    .required("Version is required"),
  compositeProfile: Yup.string().required(),
  overrideValues: Yup.array()
    .of(Yup.object())
    .typeError("Invalid override values, expected array"),
});

function DigFormGeneral(props) {
  const { item, onSubmit } = props;
  const [selectedAppIndex, setSelectedAppIndex] = useState(0); //let the first composite app as default selection
  useEffect(() => {
    if (item) {
      props.data.compositeApps.forEach((ca, index) => {
        if (ca.metadata.name === item.compositeApp) {
          setSelectedAppIndex(index);
        }
      });
    }
  }, []);

  let initialValues = item
    ? {
        ...item,
      }
    : {
        name: "",
        description: "",
        overrideValues: undefined,
        compositeApp: props.data.compositeApps[selectedAppIndex].metadata.name,
        compositeProfile: "",
        version: "",
      };

  const handleSetCompositeApp = (val) => {
    props.data.compositeApps.forEach((ca, index) => {
      if (ca.metadata.name === val) setSelectedAppIndex(index);
    });
  };
  return (
    <Formik
      initialValues={initialValues}
      onSubmit={(values) => {
        values.compositeAppVersion =
          props.data.compositeApps[selectedAppIndex].spec.version;
        onSubmit(values);
      }}
      validationSchema={schema}
    >
      {(formicProps) => {
        const {
          values,
          touched,
          errors,
          isSubmitting,
          handleChange,
          handleBlur,
          handleSubmit,
        } = formicProps;
        return (
          <form noValidate onSubmit={handleSubmit} onChange={handleChange}>
            <Grid container spacing={4} justify="center">
              <Grid container item xs={12} spacing={8}>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    id="name"
                    label="Name"
                    type="text"
                    value={values.name}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    helperText={
                      errors.name && touched.name && "Name is required"
                    }
                    required
                    error={errors.name && touched.name}
                  />
                </Grid>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    id="version"
                    label="Version"
                    type="text"
                    name="version"
                    value={values.version}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    helperText={
                      errors.version && touched.version && errors["version"]
                    }
                    required
                    error={errors.version && touched.version}
                  />
                </Grid>
              </Grid>

              <Grid item container xs={12} spacing={8}>
                <Grid item xs={12} md={6}>
                  <InputLabel shrink htmlFor="compositeApp-label-placeholder">
                    Composite App
                  </InputLabel>
                  <Select
                    fullWidth
                    name="compositeApp"
                    value={values.compositeApp}
                    onChange={(e) => {
                      handleChange(e);
                      handleSetCompositeApp(e.target.value);
                    }}
                    onBlur={handleBlur}
                    inputProps={{
                      name: "compositeApp",
                      id: "compositeApps-label-placeholder",
                    }}
                  >
                    {props.data &&
                      props.data.compositeApps.map((compositeApp) => (
                        <MenuItem
                          value={compositeApp.metadata.name}
                          key={compositeApp.metadata.name}
                        >
                          {compositeApp.metadata.name}
                        </MenuItem>
                      ))}
                  </Select>
                </Grid>
                <Grid item xs={12} md={6}>
                  <FormControl
                    fullWidth
                    required
                    error={errors.compositeProfile && touched.compositeProfile}
                  >
                    <InputLabel htmlFor="compositeProfile-label-placeholder">
                      Composite Profile
                    </InputLabel>
                    <Select
                      name="compositeProfile"
                      onChange={handleChange}
                      onBlur={handleBlur}
                      required
                      value={values.compositeProfile}
                      inputProps={{
                        name: "compositeProfile",
                        id: "compositeProfile-label-placeholder",
                      }}
                    >
                      {props.data.compositeApps[selectedAppIndex].profiles &&
                        props.data.compositeApps[selectedAppIndex].profiles.map(
                          (compositeProfile) => (
                            <MenuItem
                              value={compositeProfile.metadata.name}
                              key={compositeProfile.metadata.name}
                            >
                              {compositeProfile.metadata.name}
                            </MenuItem>
                          )
                        )}
                    </Select>
                    {errors.compositeProfile && touched.compositeProfile && (
                      <FormHelperText>Required</FormHelperText>
                    )}
                  </FormControl>
                </Grid>
              </Grid>

              <Grid item container xs={12} spacing={8}>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    name="description"
                    value={values.description}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    id="description"
                    label="Description"
                    multiline
                    rowsMax={4}
                  />
                </Grid>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    id="overrideValues"
                    label="Override Values"
                    type="text"
                    value={values.overrideValues}
                    onChange={handleChange}
                    onBlur={handleBlur}
                    multiline
                    rows={4}
                    variant="outlined"
                    error={errors.overrideValues && touched.overrideValues}
                    helperText={
                      errors.overrideValues &&
                      touched.overrideValues &&
                      errors["overrideValues"]
                    }
                  />
                </Grid>
              </Grid>
              <Grid item xs={12}>
                <DialogActions>
                  <Button
                    autoFocus
                    disabled
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
                    Next
                  </Button>
                </DialogActions>
              </Grid>
            </Grid>
          </form>
        );
      }}
    </Formik>
  );
}

export default DigFormGeneral;
