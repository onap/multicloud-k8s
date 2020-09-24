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
import React, { useState, useEffect } from "react";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import MuiDialogTitle from "@material-ui/core/DialogTitle";
import MuiDialogContent from "@material-ui/core/DialogContent";
import MuiDialogActions from "@material-ui/core/DialogActions";
import IconButton from "@material-ui/core/IconButton";
import CloseIcon from "@material-ui/icons/Close";
import Typography from "@material-ui/core/Typography";
import {
  TextField,
  InputLabel,
  Select,
  MenuItem,
  FormControl,
  FormHelperText,
  Grid,
} from "@material-ui/core";
import * as Yup from "yup";
import { Formik } from "formik";
import apiService from "../services/apiService";

const styles = (theme) => ({
  root: {
    margin: 0,
    padding: theme.spacing(2),
  },
  closeButton: {
    position: "absolute",
    right: theme.spacing(1),
    top: theme.spacing(1),
    color: theme.palette.grey[500],
  },
});
const DialogTitle = withStyles(styles)((props) => {
  const { children, classes, onClose, ...other } = props;
  return (
    <MuiDialogTitle disableTypography className={classes.root} {...other}>
      <Typography variant="h6">{children}</Typography>
      {onClose ? (
        <IconButton className={classes.closeButton} onClick={onClose}>
          <CloseIcon />
        </IconButton>
      ) : null}
    </MuiDialogTitle>
  );
});

const DialogActions = withStyles((theme) => ({
  root: {
    margin: 0,
    padding: theme.spacing(1),
  },
}))(MuiDialogActions);

const DialogContent = withStyles((theme) => ({
  root: {
    padding: theme.spacing(2),
    width: "400px",
  },
}))(MuiDialogContent);

const schema = Yup.object({
  name: Yup.string().required(),
  description: Yup.string(),
  networkControllerIntent: Yup.string(),
  genericPlacementIntent: Yup.string().required("Required"),
});

const IntentsForm = (props) => {
  const { data, onClose, open, onSubmit } = props;
  const buttonLabel = "Add";
  const title = "Add Intents";
  const [networkControllerIntents, setNetworkControllerIntents] = useState([]);
  const [genericPlacementIntents, setGenericPlacementIntents] = useState([]);
  const handleClose = () => {
    onClose();
  };
  useEffect(() => {
    let request = {
      projectName: props.projectName,
      compositeAppName: data.compositeAppName,
      compositeAppVersion: data.compositeAppVersion,
    };
    apiService
      .getGenericPlacementIntents(request)
      .then((res) => {
        setGenericPlacementIntents(res);
      })
      .catch((err) => {
        console.log("error getting network generic placement intents : ", err);
      });
    apiService
      .getNetworkControllerIntents(request)
      .then((res) => {
        setNetworkControllerIntents(res);
      })
      .catch((err) => {
        console.log("error getting network controller intents : ", err);
      });
    return function cleanup() {
      setNetworkControllerIntents([]);
      setGenericPlacementIntents([]);
    };
  }, [data, props.projectName]);
  let initialValues = {
    name: "",
    description: "",
    networkControllerIntent: "",
    genericPlacementIntent: "",
  };

  return (
    <Dialog
      onClose={handleClose}
      aria-labelledby="customized-dialog-title"
      open={open}
      disableBackdropClick
    >
      <DialogTitle id="simple-dialog-title">{title}</DialogTitle>
      <Formik
        initialValues={initialValues}
        onSubmit={async (values) => {
          values.compositeAppName = data.compositeAppName;
          values.compositeAppVersion = data.compositeAppVersion;
          values.deploymentIntentGroupName = data.metadata.name;
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
              <DialogContent dividers>
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    <TextField
                      style={{ width: "100%" }}
                      id="deploymentIntentGroupName"
                      label="Deployment Intent Group"
                      type="text"
                      value={data.metadata.name}
                      InputProps={{
                        readOnly: true,
                      }}
                    />
                  </Grid>
                  <Grid item xs={6}>
                    <TextField
                      id="name"
                      label="Name"
                      type="text"
                      required
                      onBlur={handleBlur}
                      helperText={
                        errors.name && touched.name && "Name is required"
                      }
                      error={errors.name && touched.name}
                    />
                  </Grid>
                  <Grid item xs={6}>
                    <TextField
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
                  <Grid item xs={12}>
                    <FormControl
                      error={
                        errors.genericPlacementIntent &&
                        touched.genericPlacementIntent
                      }
                      required
                      style={{ width: "100%" }}
                    >
                      <InputLabel id="genericPlacementIntent-select-label">
                        Generic placement intent
                      </InputLabel>
                      <Select
                        labelId="demo-select-label"
                        id="genericPlacementIntent-select"
                        name="genericPlacementIntent"
                        value={values.genericPlacementIntent}
                        onChange={handleChange}
                        onBlur={handleBlur}
                      >
                        {genericPlacementIntents &&
                          genericPlacementIntents.map(
                            (genericPlacementIntent) => (
                              <MenuItem
                                value={genericPlacementIntent.metadata.name}
                                key={genericPlacementIntent.metadata.name}
                              >
                                {genericPlacementIntent.metadata.name}
                              </MenuItem>
                            )
                          )}
                      </Select>
                      <FormHelperText>
                        {errors.genericPlacementIntent}
                      </FormHelperText>
                    </FormControl>
                  </Grid>

                  <Grid item xs={12}>
                    <FormControl style={{ width: "100%" }}>
                      <InputLabel id="networkControllerIntent-select-label">
                        Network controller intent
                      </InputLabel>
                      <Select
                        labelId="networkControllerIntent-select-label"
                        id="networkControllerIntent-select"
                        name="networkControllerIntent"
                        value={values.networkControllerIntent}
                        onChange={handleChange}
                        onBlur={handleBlur}
                      >
                        {networkControllerIntents &&
                          networkControllerIntents.map(
                            (networkControllerIntent) => (
                              <MenuItem
                                value={networkControllerIntent.metadata.name}
                                key={networkControllerIntent.metadata.name}
                              >
                                {networkControllerIntent.metadata.name}
                              </MenuItem>
                            )
                          )}
                      </Select>
                      <FormHelperText>
                        {errors.networkControllerIntent}
                      </FormHelperText>
                    </FormControl>
                  </Grid>
                </Grid>
              </DialogContent>
              <DialogActions>
                <Button autoFocus onClick={handleClose} color="secondary">
                  Cancel
                </Button>
                <Button
                  autoFocus
                  type="submit"
                  color="primary"
                  disabled={isSubmitting}
                >
                  {buttonLabel}
                </Button>
              </DialogActions>
            </form>
          );
        }}
      </Formik>
    </Dialog>
  );
};

IntentsForm.propTypes = {};

export default IntentsForm;
