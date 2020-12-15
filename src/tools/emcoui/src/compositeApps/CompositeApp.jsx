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
import Tab from "@material-ui/core/Tab";
import Tabs from "@material-ui/core/Tabs";
import Paper from "@material-ui/core/Paper";
import { withStyles } from "@material-ui/core/styles";
import Box from "@material-ui/core/Box";
import PropTypes from "prop-types";
import BackIcon from "@material-ui/icons/ArrowBack";
import { withRouter } from "react-router-dom";
import { IconButton } from "@material-ui/core";
import apiService from "../services/apiService";
import Spinner from "../common/Spinner";
import Apps from "../compositeApps/apps/Apps";
import CompositeProfiles from "../compositeApps/compositeProfiles/CompositeProfiles";
// import Intents from "../compositeApps/intents/GenericPlacementIntents";
// import NetworkIntent from "../networkIntents/NetworkIntents";

const lightColor = "rgba(255, 255, 255, 0.7)";

function TabPanel(props) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`nav-tabpanel-${index}`}
      aria-labelledby={`nav-tab-${index}`}
      {...other}
    >
      {value === index && <Box p={3}>{children}</Box>}
    </div>
  );
}

TabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.any.isRequired,
  value: PropTypes.any.isRequired,
};

const styles = (theme) => ({
  secondaryBar: {
    zIndex: 0,
  },
  menuButton: {
    marginLeft: -theme.spacing(1),
  },
  iconButtonAvatar: {
    padding: 4,
  },
  link: {
    textDecoration: "none",
    color: lightColor,
    "&:hover": {
      color: theme.palette.common.white,
    },
  },
  button: {
    borderColor: lightColor,
  },
});

class CompositeApp extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      activeTab: 0,
      compositeAppName: this.props.match.params.appname,
      compositeAppVersion: this.props.match.params.version,
      appsData: [],
      isLoading: true,
    };
  }
  value = 2;
  setValue(newValue) {
    this.value = newValue;
    this.setState({ activeTab: newValue });
  }
  handleChange = (event, newValue) => {
    this.setValue(newValue);
  };

  componentDidMount() {
    let request = {
      projectName: this.props.projectName,
      compositeAppName: this.props.match.params.appname,
      compositeAppVersion: this.props.match.params.version,
    };
    apiService
      .getApps(request)
      .then((res) => {
        this.setState({ appsData: res });
        this.setState({ isLoading: false });
      })
      .catch((err) => {
        console.log("error getting apps");
      });
  }

  handleUpdateState = (updatedData) => {
    this.setState({ appsData: updatedData });
  };

  render() {
    return (
      <React.Fragment>
        <div style={{ paddingBottom: "20px" }}>
          <IconButton size="small" onClick={this.props.history.goBack}>
            <BackIcon color="primary"></BackIcon>
          </IconButton>
        </div>
        <Paper square>
          <Tabs
            value={this.state.activeTab}
            indicatorColor="primary"
            textColor="primary"
            onChange={this.handleChange}
            style={{ borderBottom: "1px solid #e8e8e8" }}
          >
            <Tab label="Apps" />
            <Tab label="Composite Profiles" />
          </Tabs>
          {this.state.isLoading && <Spinner />}

          {!this.state.isLoading && (
            <>
              <TabPanel value={this.state.activeTab} index={0}>
                <Apps
                  projectName={this.props.projectName}
                  compositeAppName={this.state.compositeAppName}
                  compositeAppVersion={this.state.compositeAppVersion}
                  data={this.state.appsData}
                  onStateChange={this.handleUpdateState}
                />
              </TabPanel>
              <TabPanel value={this.state.activeTab} index={1}>
                <CompositeProfiles
                  projectName={this.props.projectName}
                  compositeAppName={this.state.compositeAppName}
                  compositeAppVersion={this.state.compositeAppVersion}
                  appsData={this.state.appsData}
                />
              </TabPanel>
            </>
          )}
        </Paper>
      </React.Fragment>
    );
  }
}
CompositeApp.propTypes = {};
export default withStyles(styles)(withRouter(CompositeApp));
