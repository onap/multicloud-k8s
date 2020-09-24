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
import PropTypes from "prop-types";
import { ThemeProvider, withStyles } from "@material-ui/core/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import Hidden from "@material-ui/core/Hidden";
import Navigator from "./AdminNavigator";
import Header from "../appbase/Header";
import theme from "../theme/Theme";
import Projects from "./projects/Projects";
import ClusterProviders from "./clusterProvider/ClusterProviders";
import Controllers from "./controllers/Controllers";

import { Switch, Route } from "react-router-dom";

const drawerWidth = 256;
const styles = {
  root: {
    display: "flex",
    minHeight: "100vh",
  },
  drawer: {
    [theme.breakpoints.up("sm")]: {
      width: drawerWidth,
      flexShrink: 0,
    },
  },
  app: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
  },
  main: {
    flex: 1,
    padding: theme.spacing(6, 4),
    background: "#eaeff1",
  },
  footer: {
    background: "#eaeff1",
  },
};

class Admin extends React.Component {
  constructor(props) {
    super(props);
    this.state = { mobileOpen: false };
  }

  setMobileOpen = (mobileOpen) => {
    this.setState({ mobileOpen });
  };
  handleDrawerToggle = () => {
    this.setMobileOpen(!this.state.mobileOpen);
  };

  render() {
    const { classes } = this.props;
    return (
      <ThemeProvider theme={theme}>
        <div className={classes.root}>
          <CssBaseline />
          <nav className={classes.drawer}>
            <Hidden smUp implementation="js">
              <Navigator
                PaperProps={{ style: { width: drawerWidth } }}
                variant="temporary"
                open={this.state.mobileOpen}
                onClose={this.handleDrawerToggle}
              />
            </Hidden>
            <Hidden xsDown implementation="css">
              <Navigator
                PaperProps={{ style: { width: drawerWidth } }}
                variant="permanent"
              />
            </Hidden>
          </nav>
          <div className={classes.app}>
            <Header onDrawerToggle={this.handleDrawerToggle} />
            <main className={classes.main}>
              <Switch>
                <Route
                  path={`${this.props.match.url}/projects`}
                  component={Projects}
                />
                <Route
                  path={`${this.props.match.url}/clusters`}
                  component={ClusterProviders}
                />
                <Route
                  path={`${this.props.match.url}/controllers`}
                  component={Controllers}
                />
              </Switch>
            </main>
          </div>
        </div>
      </ThemeProvider>
    );
  }
}

Admin.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(Admin);
