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
import ProjectsTable from "./ProjectsTable"
import { withStyles, Button, Grid } from "@material-ui/core";
import AddIcon from "@material-ui/icons/Add";
import apiService from "../../services/apiService"
import Spinner from "../../common/Spinner"
import ProjectForm from "./ProjectForm"

const styles = {
    root: {
        display: "flex",
        minHeight: "100vh",
    },
    app: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
    },
};

const Projects = () => {
    const [open, setOpen] = React.useState(false);
    const [projectsData, setProjectsData] = useState([]);
    const [isLoading, setIsloading] = useState(true);
    const handleClose = () => {
        setOpen(false);
    };
    const onCreateProject = () => {
        setOpen(true);
    }
    const handleSubmit = (updatedFields) => {
        let payload = { "metadata": updatedFields };
        apiService.createProject(payload).then(response => {
            if (projectsData && projectsData.length > 0)
                setProjectsData([...projectsData, response])
            else
                setProjectsData([response])
        }).catch(error => {
            console.log("error creating project : ", error);
        }).finally(() => {
            setOpen(false);
        })
    };
    useEffect(() => {
        apiService.getAllProjects().then(response => {
            setProjectsData(response);
        }).catch(error => {
            console.log(error)
        }).finally(() => { setIsloading(false); })
    }, []);

    return (
        <>
            {isLoading && (<Spinner />)}
            {!isLoading && (
                <>
                    <Button variant="outlined" color="primary" startIcon={<AddIcon />} onClick={onCreateProject}>
                        Create Project
                    </Button>
                    <ProjectForm open={open} onClose={handleClose} onSubmit={handleSubmit} />
                    <Grid container spacing={2} alignItems="center">
                        <Grid item xs style={{ marginTop: "20px" }}>
                            <ProjectsTable data={projectsData} setProjectsData={setProjectsData} />
                        </Grid>
                    </Grid>
                </>)}
        </>
    );
}
export default withStyles(styles)(Projects);
