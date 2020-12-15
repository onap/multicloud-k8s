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
import AppBar from "@material-ui/core/AppBar";
import Grid from "@material-ui/core/Grid";
import Hidden from "@material-ui/core/Hidden";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/core/styles";
import { withRouter } from "react-router-dom";

const lightColor = "rgba(255, 255, 255, 0.7)";

const styles = (theme) => ({
  root: {
    boxShadow:
      "0 3px 4px 0 rgba(0,0,0,.2), 0 3px 3px -2px rgba(0,0,0,.14), 0 1px 8px 0 rgba(0,0,0,.12)",
  },
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

function Header(props) {
  const { classes, onDrawerToggle, location } = props;

  let headerName = "";
  let getHeaderName = () => {
    if (location.pathname === `${props.match.url}/dashboard`) {
      headerName = "Dashboard";
    } else if (location.pathname === `${props.match.url}/services`) {
      headerName = "Services";
    } else if (
      location.pathname === `${props.match.url}/deployment-intent-group`
    ) {
      headerName = "Deployment Intent Groups";
    } else if (location.pathname.includes("services")) {
      headerName =
        "services / " +
        location.pathname.slice(location.pathname.indexOf("services")).slice(9);
    } else if (location.pathname === `${props.match.url}/projects`) {
      headerName = "Projects";
    } else if (location.pathname === `${props.match.url}/clusters`) {
      headerName = "Clusters";
    } else if (location.pathname === `${props.match.url}/controllers`) {
      headerName = "Controllers";
    }
  };
  getHeaderName();
  return (
    <React.Fragment>
      <AppBar
        className={classes.root}
        color="primary"
        position="sticky"
        elevation={0}
      >
        <Toolbar>
          <Grid container spacing={1} alignItems="center">
            <Hidden smUp implementation="js">
              <Grid item>
                <IconButton
                  color="inherit"
                  onClick={onDrawerToggle}
                  className={classes.menuButton}
                >
                  <MenuIcon />
                </IconButton>
              </Grid>
            </Hidden>
            <Typography>{headerName}</Typography>
            <Grid item xs />
          </Grid>
        </Toolbar>
      </AppBar>
    </React.Fragment>
  );
}

Header.propTypes = {
  classes: PropTypes.object.isRequired,
  onDrawerToggle: PropTypes.func.isRequired,
};

export default withStyles(styles)(withRouter(Header));
