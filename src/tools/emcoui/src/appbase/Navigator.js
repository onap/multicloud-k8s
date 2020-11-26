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
import PropTypes from "prop-types";
import clsx from "clsx";
import { withStyles } from "@material-ui/core/styles";
import Divider from "@material-ui/core/Divider";
import Drawer from "@material-ui/core/Drawer";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import HomeIcon from "@material-ui/icons/Home";
import DeviceHubIcon from "@material-ui/icons/DeviceHub";
import DnsRoundedIcon from "@material-ui/icons/DnsRounded";
import { withRouter, Link } from "react-router-dom";

const categories = [
  {
    id: "1",
    children: [
      {
        id: "Services",
        icon: <DeviceHubIcon />,
        url: "/services",
      },
      {
        id: "Deployment Intent Groups",
        icon: <DnsRoundedIcon />,
        url: "/deployment-intent-group",
      },
    ],
  },
];

const styles = (theme) => ({
  categoryHeader: {
    paddingTop: theme.spacing(2),
    paddingBottom: theme.spacing(2),
  },
  categoryHeaderPrimary: {
    color: theme.palette.common.white,
  },
  item: {
    paddingTop: 1,
    paddingBottom: 1,
    color: "rgba(255, 255, 255, 0.7)",
    "&:hover,&:focus": {
      backgroundColor: "rgba(255, 255, 255, 0.08)",
    },
  },
  itemCategory: {
    backgroundColor: "#232f3e",
    boxShadow: "0 -1px 0 #404854 inset",
    paddingTop: theme.spacing(2),
    paddingBottom: theme.spacing(2),
  },
  firebase: {
    fontSize: 24,
    color: theme.palette.common.white,
  },
  itemActiveItem: {
    color: theme.palette.primary.main,
  },
  itemPrimary: {
    fontSize: "inherit",
  },
  itemIcon: {
    minWidth: "auto",
    marginRight: theme.spacing(2),
  },
  divider: {
    marginTop: theme.spacing(2),
  },
  version: {
    fontSize: "15px",
  },
});

function Navigator(props) {
  const { classes, location } = props;
  const [activeItem, setActiveItem] = useState(location.pathname);
  const setActiveTab = (itemId) => {
    setActiveItem(itemId);
  };
  if (location.pathname !== activeItem) {
    setActiveTab(location.pathname);
  }
  return (
    <Drawer
      PaperProps={props.PaperProps}
      variant={props.variant}
      open={props.open}
      onClose={props.onClose}
    >
      <List disablePadding>
        <Link style={{ textDecoration: "none" }} to="/">
          <ListItem
            className={clsx(
              classes.firebase,
              classes.item,
              classes.itemCategory
            )}
          >
            <ListItemText
              classes={{
                primary: classes.itemPrimary,
              }}
            >
              ONAP4K8s
            </ListItemText>
            <span className={clsx(classes.version)}>
              {process.env.REACT_APP_VERSION}
            </span>
          </ListItem>
        </Link>

        {/* <Link
          style={{ textDecoration: "none" }}
          to={{
            pathname: `${props.match.url}/dashboard`,
            activeItem: "childId",
          }}
          key={"childId"}
        > */}
        <ListItem
          button
          className={clsx(
            classes.item,
            classes.itemCategory,
            activeItem.includes("dashboard") && classes.itemActiveItem
          )}
        >
          <ListItemIcon className={classes.itemIcon}>
            <HomeIcon />
          </ListItemIcon>
          <ListItemText
            classes={{
              primary: classes.itemPrimary,
            }}
          >
            Dashboard
          </ListItemText>
        </ListItem>
        {/* </Link> */}
        {categories.map(({ id, children }) => (
          <React.Fragment key={id}>
            {children.map(({ id: childId, icon, url }) => (
              <Link
                style={{ textDecoration: "none" }}
                to={{
                  pathname: `${props.match.url}${url}`,
                  activeItem: childId,
                }}
                key={childId}
              >
                <ListItem
                  button
                  className={clsx(
                    classes.item,
                    activeItem.includes(url) && classes.itemActiveItem
                  )}
                >
                  <ListItemIcon className={classes.itemIcon}>
                    {icon}
                  </ListItemIcon>
                  <ListItemText
                    classes={{
                      primary: classes.itemPrimary,
                    }}
                  >
                    {childId}
                  </ListItemText>
                </ListItem>
              </Link>
            ))}

            <Divider className={classes.divider} />
          </React.Fragment>
        ))}
      </List>
    </Drawer>
  );
}

Navigator.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(withRouter(Navigator));
