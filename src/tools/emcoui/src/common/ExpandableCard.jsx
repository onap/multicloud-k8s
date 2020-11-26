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
import clsx from "clsx";
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Collapse from "@material-ui/core/Collapse";
import IconButton from "@material-ui/core/IconButton";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import StorageIcon from "@material-ui/icons/Storage";
import ErrorIcon from "@material-ui/icons/Error";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
  },
  expand: {
    transform: "rotate(0deg)",
    marginLeft: "auto",
    transition: theme.transitions.create("transform", {
      duration: theme.transitions.duration.shortest,
    }),
  },
  expandOpen: {
    transform: "rotate(180deg)",
  },
}));
const ExpandableCard = (props) => {
  const classes = useStyles();
  const [expanded, setExpanded] = useState(false);

  const handleExpandClick = () => {
    if (!expanded) {
      setExpanded(!expanded);
    } else {
      setExpanded(!expanded);
    }
  };

  return (
    <>
      <Card className={classes.root}>
        <CardHeader
          onClick={handleExpandClick}
          avatar={
            <>
              <StorageIcon fontSize="large" />
            </>
          }
          action={
            <>
              {props.error && (
                <ErrorIcon color="error" style={{ verticalAlign: "middle" }} />
              )}
              <IconButton
                className={clsx(classes.expand, {
                  [classes.expandOpen]: expanded,
                })}
                onClick={handleExpandClick}
                aria-expanded={expanded}
              >
                <ExpandMoreIcon />
              </IconButton>
            </>
          }
          title={props.title}
          subheader={props.description}
        />
        <Collapse in={expanded} timeout="auto" unmountOnExit>
          <CardContent>{props.content}</CardContent>
        </Collapse>
      </Card>
    </>
  );
};

ExpandableCard.propTypes = {};

export default ExpandableCard;
