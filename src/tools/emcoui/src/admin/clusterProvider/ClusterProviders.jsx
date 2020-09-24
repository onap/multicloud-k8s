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
import React, { useEffect, useState } from 'react';
import apiService from "../../services/apiService"
import { Button, Grid } from '@material-ui/core';
import Spinner from '../../common/Spinner';
import AddIcon from "@material-ui/icons/Add";
import ClusterProviderForm from "./ClusterProviderForm";
import ClusterProvidersAccordian from "./ClusterProvidersAccordian";

const ClusterProviders = () => {
    const [data, setData] = useState([]);
    const [isLoading, setIsloading] = useState(true);
    const [openForm, setOpenForm] = React.useState(false);

    const handleClose = () => {
        setOpenForm(false);
    };

    const handleSubmit = (updatedFields) => {
        let payload = { "metadata": updatedFields };
        apiService.registerClusterProvider(payload).then(res => {
            (!data || data.length === 0) ? setData([res]) : setData([...data, res]);
        }).catch(error => {
            console.log("error registering cluster provider : ", error);
        }).finally(() => {
            setOpenForm(false);
        })
    };
    useEffect(() => {
        apiService.getClusterProviders().then(response => {
            setData(response);
        }).catch(error => {
            console.log("error getting cluster providers : ", error)
        }).finally(() => { setIsloading(false); })
    }, []);

    const onRegisterClusterProvider = () => {
        setOpenForm(true);
    }
    return (
        <>
            {isLoading && (<Spinner />)}
            {!isLoading && (
                <>
                    <Button variant="outlined" color="primary" startIcon={<AddIcon />} onClick={onRegisterClusterProvider}>
                        Register Cluster Provider
                    </Button>
                    <ClusterProviderForm open={openForm} onClose={handleClose} onSubmit={handleSubmit} />
                    <Grid container spacing={2} alignItems="center">
                        <Grid item xs style={{ marginTop: "20px" }}>
                            <ClusterProvidersAccordian data={data} setData={setData} />
                        </Grid>
                    </Grid>
                </>)}
        </>
    );
};


ClusterProviders.propTypes = {

};


export default ClusterProviders;
