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
import Card from "./CompositeProfileCard";
import { Button, Grid } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import PropTypes from "prop-types";
import Form from "../../common/Form";
import apiService from "../../services/apiService";
import DeleteDialog from "../../common/Dialogue";

const CompositeProfiles = (props) => {
  const [openForm, setOpenForm] = useState(false);
  const [data, setData] = useState([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [index, setIndex] = useState(0);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let request = {
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
    };
    apiService
      .getCompositeProfiles(request)
      .then((res) => {
        setData(res);
      })
      .catch((err) => {
        console.log("error geting composite profiles ", err);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [props.projectName, props.compositeAppName, props.compositeAppVersion]);
  const handleCloseForm = () => {
    setOpenForm(false);
  };
  const handleAddCompositeProfile = () => {
    setOpenForm(true);
  };
  const handleSubmit = (values) => {
    let request = {
      payload: {
        metadata: { name: values.name, description: values.description },
      },
      projectName: props.projectName,
      compositeAppName: props.compositeAppName,
      compositeAppVersion: props.compositeAppVersion,
    };
    apiService
      .createCompositeProfile(request)
      .then((res) => {
        !data || data.length === 0 ? setData([res]) : setData([...data, res]);
      })
      .catch((err) => {
        console.log("error creating composite profile : ", err);
      })
      .finally(() => {
        setOpenForm(false);
      });
  };

  const handleCloseDialog = (el) => {
    if (el.target.innerText === "Delete") {
      let request = {
        projectName: props.projectName,
        compositeAppName: props.compositeAppName,
        compositeAppVersion: props.compositeAppVersion,
        compositeProfileName: data[index].metadata.name,
      };
      apiService
        .deleteCompositeProfile(request)
        .then(() => {
          console.log("comsposite profile deleted");
          data.splice(index, 1);
          setData([...data]);
        })
        .catch((err) => {
          console.log("Error deleting comsposite profile : " + err);
        })
        .finally(() => {
          setIndex(0);
        });
    }
    setOpenDialog(false);
  };
  const handleDeleteCompositeProfile = (index) => {
    setIndex(index);
    setOpenDialog(true);
  };

  return (
    <>
      <DeleteDialog
        open={openDialog}
        onClose={handleCloseDialog}
        title={"Delete Composite Profile"}
        content={`Are you sure you want to delete "${
          data && data[index] ? data[index].metadata.name : ""
        }"`}
      />

      {/* <Button
        disabled={isLoading}
        variant="outlined"
        color="primary"
        startIcon={<AddIcon />}
        onClick={handleAddCompositeProfile}
      >
        Add Composite Profile
      </Button> */}
      <Form onClose={handleCloseForm} open={openForm} onSubmit={handleSubmit} />
      <Grid
        container
        justify="flex-start"
        style={{ display: "flex", marginTop: "15px" }}
      >
        {data &&
          data.map((compositeProfile, compositeProfileIndex) => (
            <Card
              key={compositeProfile.metadata.name}
              compositeProfile={compositeProfile}
              projectName={props.projectName}
              compositeAppName={props.compositeAppName}
              compositeAppVersion={props.compositeAppVersion}
              appsData={props.appsData}
              index={compositeProfileIndex}
              onDeleteCompositeProfile={handleDeleteCompositeProfile}
            />
          ))}
      </Grid>
    </>
  );
};

CompositeProfiles.propTypes = {
  projectName: PropTypes.string.isRequired,
  compositeAppName: PropTypes.string.isRequired,
};

export default CompositeProfiles;
