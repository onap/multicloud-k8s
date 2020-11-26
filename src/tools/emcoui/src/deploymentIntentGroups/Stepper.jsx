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
import { makeStyles } from "@material-ui/core/styles";
import Stepper from "@material-ui/core/Stepper";
import Step from "@material-ui/core/Step";
import StepLabel from "@material-ui/core/StepLabel";
import DigFormGeneral from "./DigFormGeneral";
import DigFormIntents from "./DigFormIntents";
import apiService from "../services/apiService";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
  },
  backButton: {
    marginRight: theme.spacing(1),
  },
  instructions: {
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
  },
}));

function getSteps() {
  return ["General", "Intents"];
}

export default function HorizontalStepper(props) {
  const classes = useStyles();
  const [activeStep, setActiveStep] = useState(0);
  const [generalData, setGeneralData] = useState(null);
  const [intentsData, setIntentsData] = useState(null);
  const [appsData, setAppsData] = useState([]);

  const steps = getSteps();

  function getStepContent(stepIndex) {
    switch (stepIndex) {
      case 0:
        return (
          <DigFormGeneral
            data={props.data}
            onSubmit={handleGeneralFormSubmit}
            item={generalData}
          />
        );
      case 1:
        return (
          <DigFormIntents
            appsData={appsData}
            onSubmit={handleIntentsFormSubmit}
            onClickBack={handleBack}
            item={intentsData}
          />
        );
      default:
        return "Unknown stepIndex";
    }
  }

  const handleNext = () => {
    setActiveStep((prevActiveStep) => prevActiveStep + 1);
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
  };
  const handleGeneralFormSubmit = (values) => {
    setGeneralData(values);
    let request = {
      projectName: props.projectName,
      compositeAppName: values.compositeApp,
      compositeAppVersion: values.compositeAppVersion,
    };
    apiService
      .getApps(request)
      .then((res) => {
        setAppsData(res);
        handleNext((prevActiveStep) => prevActiveStep + 1);
      })
      .catch((err) => {
        console.log("Error getting apps : " + err);
      });
  };

  const handleIntentsFormSubmit = (values) => {
    setIntentsData(values);
    let digPayload = { general: generalData, intents: values };
    props.onSubmit(digPayload);
  };
  return (
    <div className={classes.root}>
      <Stepper activeStep={activeStep} alternativeLabel>
        {steps.map((label) => (
          <Step key={label}>
            <StepLabel>{label}</StepLabel>
          </Step>
        ))}
      </Stepper>
      <div>
        <div>{getStepContent(activeStep)}</div>
      </div>
    </div>
  );
}
