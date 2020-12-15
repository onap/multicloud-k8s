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
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import MuiDialogTitle from "@material-ui/core/DialogTitle";
import MuiDialogContent from "@material-ui/core/DialogContent";
import MuiDialogActions from "@material-ui/core/DialogActions";
import IconButton from "@material-ui/core/IconButton";
import CloseIcon from "@material-ui/icons/Close";
import Typography from "@material-ui/core/Typography";
import Stepper from "./Stepper";
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
  },
}))(MuiDialogContent);

const DIGform = (props) => {
  const { onClose, item, open, onSubmit } = props;
  const title = item
    ? "Edit Deployment Intent Group"
    : "Create Deployment Intent Group";
  const handleClose = () => {
    onClose();
  };
  useEffect(() => {
    props.data.compositeApps.forEach((compositeApp) => {
      let request = {
        projectName: props.projectName,
        compositeAppName: compositeApp.metadata.name,
        compositeAppVersion: compositeApp.spec.version,
      };
      apiService
        .getCompositeProfiles(request)
        .then((res) => {
          compositeApp.profiles = res;
        })
        .catch((error) => {
          console.log("error getting cluster providers : ", error);
        })
        .finally(() => {});
    });
  }, [props.data.compositeApps, props.projectName]);
  return (
    <Dialog
      maxWidth={"md"}
      fullWidth={true}
      onClose={handleClose}
      open={open}
      disableBackdropClick
    >
      <DialogTitle id="customized-dialog-title" onClose={handleClose}>
        {title}
      </DialogTitle>
      <DialogContent dividers>
        <Stepper
          data={props.data}
          projectName={props.projectName}
          onSubmit={onSubmit}
        />
      </DialogContent>
    </Dialog>
  );
};

DIGform.propTypes = {
  onClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
};

export default DIGform;
