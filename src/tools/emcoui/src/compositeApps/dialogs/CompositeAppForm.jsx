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
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import MuiDialogTitle from "@material-ui/core/DialogTitle";
import MuiDialogContent from "@material-ui/core/DialogContent";
import MuiDialogActions from "@material-ui/core/DialogActions";
import Typography from "@material-ui/core/Typography";
import { TextField } from '@material-ui/core';
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


const DialogContent = withStyles((theme) => ({
  root: {
    padding: theme.spacing(2),
  },
}))(MuiDialogContent);

const DialogActions = withStyles((theme) => ({
  root: {
    margin: 0,
    padding: theme.spacing(1),
  },
}))(MuiDialogActions);


class CreateCompositeAppForm extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      fields: { name: "", version: "", description: "" },
      errors: {}
    }
    this.handleChange = this.handleChange.bind(this);
    this.submituserRegistrationForm = this.submituserRegistrationForm.bind(this);
  }


  componentDidMount = () => {
    if (this.props.item) {
      this.title = "Edit Composite App";
      this.buttonLabel = "Update";
      this.isEdit = true;
    }
    else {
      this.title = "New Composite App";
      this.buttonLabel = "Create";
      this.isEdit = false;
    }
  };

  componentDidUpdate = (prevProps, prevState) => {
    if (this.props.item && ((prevProps.item !== this.props.item))) {
      this.setState({ fields: { ...this.props.item.metadata, version: this.props.item.spec.version } });
    }
  }

  resetFields = () => {
    if (!this.isEdit) {
      this.setState({
        fields: { name: "", version: "", description: "" },
        errors: {}
      });
    }
    else {
      this.setState({ fields: { ...this.props.item.metadata, version: this.props.item.spec.version } });
    }
  }

  handleClose = () => {
    this.resetFields();
    this.props.handleClose();
  };

  submituserRegistrationForm(e) {
    e.preventDefault();
    if (this.validateForm()) {
      this.resetFields();
      this.props.handleClose(this.state.fields);
    }
  }

  validateForm() {
    let fields = this.state.fields;
    let errors = {};
    let formIsValid = true;

    if (!fields["name"]) {
      formIsValid = false;
      errors["name"] = "*Please enter your username.";
    }

    if (typeof fields["name"] !== "string") {
      if (!fields["name"].match(/^[a-zA-Z ]*$/)) {
        formIsValid = false;
        errors["name"] = "*Please enter alphabet characters only.";
      }
    }
    this.setState({
      errors: errors
    });
    return formIsValid;
  }

  handleChange = (e) => {
    this.setState({ fields: { ...this.state.fields, [e.target.name]: e.target.value } });
  }

  render = () => {
    const { classes } = this.props;
    return (
      <>
        <Dialog
          maxWidth={"xs"}
          onClose={this.handleClose}
          aria-labelledby="customized-dialog-title"
          open={this.props.open}
          disableBackdropClick
        >
          <MuiDialogTitle disableTypography className={classes.root} >
            <Typography variant="h6">{this.title}</Typography>
          </MuiDialogTitle>

          <form onSubmit={this.submituserRegistrationForm}>
            <DialogContent dividers>
              <TextField
                style={{ width: "40%", marginBottom: "10px" }}
                name="name"
                value={this.state.fields.name}
                id="input-name"
                label="Name"
                helperText="Name should be unique"
                onChange={this.handleChange}
                required
              />
              <TextField
                style={{ width: "40%", marginBottom: "20px", float: "right" }}
                name="version"
                value={this.state.fields.version}
                onChange={this.handleChange}
                id="input-version"
                label="Version"
                required
              />
              <TextField
                style={{ width: "100%", marginBottom: "25px" }}
                name="description"
                value={this.state.fields.description}
                onChange={this.handleChange}
                id="input-description"
                label="Description"
                multiline
                rowsMax={4}
              />
            </DialogContent>
            <DialogActions>
              <Button autoFocus onClick={this.handleClose} color="secondary">
                Cancel
          </Button>
              <Button autoFocus type="submit" color="primary">
                {this.buttonLabel}
              </Button>
            </DialogActions>
          </form>
        </Dialog>
      </>
    );
  }
}
export default withStyles(styles)(CreateCompositeAppForm)
