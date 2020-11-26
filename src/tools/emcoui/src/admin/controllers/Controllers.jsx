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
import React, { useEffect, useState } from "react";
import ControllersTable from "./ControllersTable";
import { Button, Grid } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../../services/apiService";
import Spinner from "../../common/Spinner";
import ControllerForm from "./ControllerForm";

function Controllers() {
  const [isLoading, setIsLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [controllersData, setControllersData] = useState([]);

  useEffect(() => {
    apiService
      .getControllers()
      .then((res) => {
        if (res && res.length > 0) setControllersData(res);
        else setControllersData([]);
      })
      .catch((err) => {
        console.log("error getting controllers : " + err);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);
  const handleClose = (values) => {
    if (values) {
      let request = {
        metadata: { name: values.name, description: values.description },
        spec: {
          host: values.host,
          port: parseInt(values.port),
        },
      };
      if (values.type) request.spec.type = values.type;
      if (values.priority) request.spec.priority = parseInt(values.priority);
      apiService
        .addController(request)
        .then((res) => {
          setControllersData((controllersData) => {
            return [...controllersData, res];
          });
        })
        .catch((err) => {
          console.log("error adding controller : " + err);
        });
    }
    setOpen(false);
  };
  const onAddController = () => {
    setOpen(true);
  };

  return (
    <>
      {isLoading && <Spinner />}
      {!isLoading && (
        <>
          <Button
            variant="outlined"
            color="primary"
            startIcon={<AddIcon />}
            onClick={onAddController}
          >
            Register Controller
          </Button>
          <ControllerForm open={open} onClose={handleClose} />
          <Grid container spacing={2} alignItems="center">
            <Grid item xs style={{ marginTop: "20px" }}>
              <ControllersTable
                data={controllersData}
                setControllerData={setControllersData}
              />
            </Grid>
          </Grid>
        </>
      )}
    </>
  );
}

export default Controllers;
