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
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Redirect,
} from "react-router-dom";
import "./App.css";
import AppBase from "./appbase/AppBase";
import Admin from "./admin/Admin";

function App() {
  return (
    <Router>
      <Switch>
        <Route
          path="/app/admin"
          children={({ match, ...others }) => {
            return (
              <Switch>
                <Redirect
                  exact
                  from={`${match.path}`}
                  to={`${match.path}/projects`}
                />
                <Route
                  path={`${match.path}`}
                  render={(props) => <Admin {...props} />}
                />
              </Switch>
            );
          }}
        />
        <Route
          path="/app/projects/:projectName"
          children={({ match, ...others }) => {
            return (
              <Switch>
                <Redirect
                  exact
                  from={`${match.path}`}
                  to={`${match.path}/services`}
                />
                <Route
                  path={`${match.path}`}
                  render={(props) => <AppBase {...props} />}
                />
              </Switch>
            );
          }}
        />
        <Route
          path="/"
          render={() => {
            return <Redirect path="/" to={"/app/admin"} />;
          }}
        />
      </Switch>
    </Router>
  );
}
export default App;
