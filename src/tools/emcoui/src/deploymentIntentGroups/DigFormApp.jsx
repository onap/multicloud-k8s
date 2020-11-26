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
import { makeStyles } from "@material-ui/core/styles";
import PropTypes from "prop-types";
import Tabs from "@material-ui/core/Tabs";
import Tab from "@material-ui/core/Tab";
import Box from "@material-ui/core/Box";
import React, { useState } from "react";
import Typography from "@material-ui/core/Typography";
import { Formik } from "formik";
import ExpandableCard from "../common/ExpandableCard";
import AppPlacementForm from "../compositeApps/dialogs/AppFormPlacement";
import NetworkForm from "../compositeApps/dialogs/AppNetworkForm";

const useStyles = makeStyles((theme) => ({
  tableRoot: {
    width: "100%",
  },
  paper: {
    width: "100%",
    marginBottom: theme.spacing(2),
  },
  table: {
    minWidth: 550,
  },
  visuallyHidden: {
    border: 0,
    clip: "rect(0 0 0 0)",
    height: 1,
    margin: -1,
    overflow: "hidden",
    padding: 0,
    position: "absolute",
    top: 20,
    width: 1,
  },
  appBar: {
    position: "relative",
  },
  title: {
    marginLeft: theme.spacing(2),
    flex: 1,
  },
  demo: {
    backgroundColor: theme.palette.background.paper,
  },
  root: {
    flexGrow: 1,
    backgroundColor: theme.palette.background.paper,
    display: "flex",
    height: 424,
  },
  tabs: {
    borderRight: `1px solid ${theme.palette.divider}`,
  },
}));
function TabPanel(props) {
  const { children, value, index, ...other } = props;
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`vertical-tabpanel-${index}`}
      aria-labelledby={`vertical-tab-${index}`}
      {...other}
    >
      {value === index && <Box style={{ padding: "0 24px" }}>{children}</Box>}
    </div>
  );
}

function AppDetailsForm({ formikProps, ...props }) {
  const classes = useStyles();
  const [value, setValue] = useState(0);
  const handleChange = (event, newValue) => {
    setValue(newValue);
  };
  const handleRowSelect = (clusterProvider, selectedClusters) => {
    if (
      !formikProps.values.apps[props.index].clusters ||
      formikProps.values.apps[props.index].clusters === undefined
    ) {
      if (selectedClusters.length > 0) {
        let selectedClusterData = [];
        selectedClusters.forEach((selectedCluster) => {
          selectedClusterData.push({ name: selectedCluster, interfaces: [] });
        });
        formikProps.setFieldValue(`apps[${props.index}].clusters`, [
          {
            provider: clusterProvider,
            selectedClusters: selectedClusterData,
          },
        ]);
      }
    } else {
      let selectedClusterData = [];
      //filter out the value of cluster provider so that it can be completely replaced by the new values
      let updatedClusterValues = formikProps.values.apps[
        props.index
      ].clusters.filter((cluster) => cluster.provider !== clusterProvider);
      selectedClusters.forEach((selectedCluster) => {
        selectedClusterData.push({ name: selectedCluster, interfaces: [] });
      });
      if (selectedClusters.length > 0)
        updatedClusterValues.push({
          provider: clusterProvider,
          selectedClusters: selectedClusterData,
        });
      formikProps.setFieldValue(
        `apps[${props.index}].clusters`,
        updatedClusterValues
      );
    }
  };
  return (
    <div className={classes.root}>
      <Formik>
        {() => {
          return (
            <>
              <Tabs
                orientation="vertical"
                variant="scrollable"
                value={value}
                onChange={handleChange}
                aria-label="Vertical tabs example"
                className={classes.tabs}
              >
                <Tab label="Placement" {...a11yProps(1)} />
                <Tab label="Network" {...a11yProps(2)} />
              </Tabs>
              <TabPanel style={{ width: "85%" }} value={value} index={0}>
                <AppPlacementForm
                  formikProps={formikProps}
                  index={props.index}
                  clusterProviders={props.clusterProviders}
                  handleRowSelect={handleRowSelect}
                />
              </TabPanel>
              <TabPanel style={{ width: "85%" }} value={value} index={1}>
                <Typography variant="subtitle1">Select Network</Typography>
                <NetworkForm
                  clusters={formikProps.values.apps[props.index].clusters}
                  formikProps={formikProps}
                  index={props.index}
                />
              </TabPanel>
            </>
          );
        }}
      </Formik>
    </div>
  );
}

TabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.any.isRequired,
  value: PropTypes.any.isRequired,
};

function a11yProps(index) {
  return {
    id: `vertical-tab-${index}`,
    "aria-controls": `vertical-tabpanel-${index}`,
  };
}

const AppForm2 = (props) => {
  return (
    <ExpandableCard
      error={
        props.formikProps.errors.apps &&
        props.formikProps.errors.apps[props.index]
      }
      title={props.name}
      description={props.description}
      content={
        <AppDetailsForm
          formikProps={props.formikProps}
          name={props.name}
          index={props.index}
          clusterProviders={props.clusterProviders}
        />
      }
    />
  );
};
export default AppForm2;
