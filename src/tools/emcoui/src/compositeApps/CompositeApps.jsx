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
import CompositeAppTable from "./CompositeAppTable";
import { withStyles, Button, Grid } from "@material-ui/core";
import CreateCompositeAppForm from "./dialogs/CompositeAppForm";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../services/apiService";
import Spinner from "../common/Spinner";
const styles = {
  root: {
    display: "flex",
    minHeight: "100vh",
  },
  app: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
  },
};

class CompositeApps extends React.Component {
  constructor(props) {
    super(props);
    this.state = { open: false, data: [], isLoading: true };
  }

  componentDidMount() {
    apiService
      .getCompositeApps({ projectName: this.props.projectName })
      .then((response) => {
        this.setState({ data: response });
      })
      .catch((err) => {
        console.log("Unable to get composite apps : ", err);
      })
      .finally(() => {
        this.setState({ isLoading: false });
      });
  }

  handleCreateCompositeApp = (row) => {
    this.setState({ open: true });
  };

  handleClose = (fields) => {
    if (fields) {
      let request = {
        payload: {
          metadata: { name: fields.name, description: fields.description },
          spec: { version: fields.version },
        },
        projectName: this.props.projectName,
      };
      apiService
        .createCompositeApp(request)
        .then((res) => {
          if (this.state.data && this.state.data.length > 0)
            this.setState({ data: [...this.state.data, res] });
          else this.setState({ data: [res] });
        })
        .catch((err) => {
          console.log("error creating composite app : ", err);
        });
    }
    this.setState({ open: false });
  };

  handleUpdateState = (updatedData) => {
    this.setState({ data: updatedData });
  };

  render = () => {
    return (
      <>
        {this.state.isLoading && <Spinner />}
        {!this.state.isLoading && (
          <>
            <Button
              variant="outlined"
              color="primary"
              startIcon={<AddIcon />}
              onClick={this.handleCreateCompositeApp}
            >
              Create Composite App
            </Button>
            <CreateCompositeAppForm
              open={this.state.open}
              handleClose={this.handleClose}
            />
            <Grid container spacing={2} alignItems="center">
              <Grid item xs style={{ marginTop: "20px" }}>
                {this.state.data && this.state.data.length > 0 && (
                  <CompositeAppTable
                    data={this.state.data}
                    projectName={this.props.projectName}
                    handleUpdateState={this.handleUpdateState}
                  />
                )}
                {(!this.state.data || this.state.data.length === 0) && (
                  <span>No Composite Apps</span>
                )}
              </Grid>
            </Grid>
          </>
        )}
      </>
    );
  };
}
export default withStyles(styles)(CompositeApps);
