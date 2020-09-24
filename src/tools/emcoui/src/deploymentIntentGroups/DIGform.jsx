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
import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import MuiDialogTitle from '@material-ui/core/DialogTitle';
import MuiDialogContent from '@material-ui/core/DialogContent';
import MuiDialogActions from '@material-ui/core/DialogActions';
import IconButton from '@material-ui/core/IconButton';
import CloseIcon from '@material-ui/icons/Close';
import Typography from '@material-ui/core/Typography';
import { TextField, InputLabel, NativeSelect, FormControl, FormHelperText } from '@material-ui/core';
import * as Yup from "yup";
import { Formik } from 'formik';
import apiService from "../services/apiService";

const styles = (theme) => ({
    root: {
        margin: 0,
        padding: theme.spacing(2),
    },
    closeButton: {
        position: 'absolute',
        right: theme.spacing(1),
        top: theme.spacing(1),
        color: theme.palette.grey[500],
    },
});

const DialogTitle = withStyles(styles)((props) => {
    const { children, classes, onClose, ...other } = props;
    return (
        <MuiDialogTitle disableTypography className={classes.root} {...other}>
            <Typography variant="h6">{children}</Typography>
            {onClose ? (
                <IconButton className={classes.closeButton} onClick={onClose}>
                    <CloseIcon />
                </IconButton>
            ) : null}
        </MuiDialogTitle>
    );
});

const DialogActions = withStyles((theme) => ({
    root: {
        margin: 0,
        padding: theme.spacing(1),
    },
}))(MuiDialogActions);

const DialogContent = withStyles((theme) => ({
    root: {
        padding: theme.spacing(2),
    }
}))(MuiDialogContent);

const schema = Yup.object(
    {
        name: Yup.string().required(),
        description: Yup.string(),
        version: Yup.string().required(),
        compositeProfile: Yup.string().required(),
        overrideValues: Yup.array().of(Yup.object()).typeError("Invalid override values, expected array"),
    })

const DIGform = (props) => {
    const { onClose, item, open, onSubmit } = props;
    const buttonLabel = item ? "OK" : "Create";
    const title = item ? "Edit Deployment Intent Group" : "Create Deployment Intent Group";
    const [selectedAppIndex, setSelectedAppIndex] = useState(0);
    const handleClose = () => {
        onClose();
    };
    useEffect(() => {
        props.data.compositeApps.forEach(compositeApp => {
            let request = { projectName: props.projectName, compositeAppName: compositeApp.metadata.name, compositeAppVersion: compositeApp.spec.version }
            apiService.getCompositeProfiles(request).then(res => {
                compositeApp.profiles = res;
            }).catch(error => {
                console.log("error getting cluster providers : ", error)
            }).finally(() => {
            })
        })
    }, [props.data.compositeApps, props.projectName]);
    let initialValues = item ?
        { name: item.metadata.name, description: item.metadata.description, overrideValues: JSON.stringify(item.spec["override-values"]), compositeApp: item.compositeAppName, compositeProfile: item.spec.profile, version: item.spec.version } :
        { name: "", description: "", overrideValues: undefined, compositeApp: props.data.compositeApps[0].metadata.name, compositeProfile: "", version: "" }

    const handleSetCompositeApp = (val) => {
        props.data.compositeApps.forEach((ca, index) => {
            if (ca.metadata.name === val)
                setSelectedAppIndex(index);
        });
    }

    return (
        <Dialog maxWidth={"xs"} onClose={handleClose} aria-labelledby="customized-dialog-title" open={open} disableBackdropClick>
            <DialogTitle id="simple-dialog-title">{title}</DialogTitle>
            <Formik
                initialValues={initialValues}
                onSubmit={async values => {
                    values.compositeAppVersion = props.data.compositeApps[selectedAppIndex].spec.version;
                    onSubmit(values);
                }}
                validationSchema={schema}
            >
                {formicProps => {
                    const {
                        values,
                        touched,
                        errors,
                        isSubmitting,
                        handleChange,
                        handleBlur,
                        handleSubmit
                    } = formicProps;
                    return (
                        <form noValidate onSubmit={handleSubmit} onChange={handleChange}>
                            <DialogContent dividers>
                                <div style={{ width: "45%", float: "left" }}>
                                    <InputLabel shrink htmlFor="compositeApp-label-placeholder">
                                        Composite App
                                    </InputLabel>
                                    <NativeSelect
                                        name="compositeApp"
                                        onChange={(e) => { handleChange(e); handleSetCompositeApp(e.target.value) }}
                                        onBlur={handleBlur}
                                        disabled={item ? true : false}
                                        inputProps={{
                                            name: 'compositeApp',
                                            id: 'compositeApps-label-placeholder',
                                        }}
                                    >
                                        {item && (<option >{values.compositeApp}</option>)}
                                        {props.data && props.data.compositeApps.map(compositeApp =>
                                            (<option value={compositeApp.metadata.name} key={compositeApp.metadata.name} >{compositeApp.metadata.name}</option>)
                                        )}
                                    </NativeSelect>
                                </div>

                                <FormControl style={{ width: "45%", float: "right" }} required error={errors.compositeProfile && touched.compositeProfile}>
                                    <InputLabel htmlFor="compositeProfile-label-placeholder">
                                        Composite Profile
                                    </InputLabel>
                                    <NativeSelect
                                        name="compositeProfile"
                                        onChange={handleChange}
                                        onBlur={handleBlur}
                                        disabled={item ? true : false}
                                        required
                                        inputProps={{
                                            name: 'compositeProfile',
                                            id: 'compositeProfile-label-placeholder',
                                        }}
                                    >
                                        <option value="" />
                                        {props.data.compositeApps[selectedAppIndex].profiles && props.data.compositeApps[selectedAppIndex].profiles.map(compositeProfile =>
                                            (<option value={compositeProfile.metadata.name} key={compositeProfile.metadata.name} >{compositeProfile.metadata.name}</option>)
                                        )}
                                    </NativeSelect>
                                    {errors.compositeProfile && touched.compositeProfile && <FormHelperText>Required</FormHelperText>}
                                </FormControl>
                                <TextField
                                    style={{ width: "45%", float: "left", marginTop: "10px" }}
                                    id="name"
                                    label="Name"
                                    type="text"
                                    value={values.name}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    helperText={(errors.name && touched.name && (
                                        "Name is required"
                                    ))}
                                    required
                                    error={errors.name && touched.name}
                                />
                                <TextField
                                    style={{ width: "45%", float: "right", marginTop: "10px" }}
                                    id="version"
                                    label="Version"
                                    type="text"
                                    name="version"
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    helperText={(errors.version && touched.version && (
                                        "Version is required"
                                    ))}
                                    required
                                    error={errors.version && touched.version}
                                />
                                <TextField
                                    style={{ width: "100%", marginTop: "20px" }}
                                    id="overrideValues"
                                    label="Override Values"
                                    type="text"
                                    value={values.overrideValues}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    required
                                    multiline
                                    rows={4}
                                    variant="outlined"
                                    error={errors.overrideValues && touched.overrideValues}
                                    helperText={(errors.overrideValues && touched.overrideValues && (
                                        (errors["overrideValues"])
                                    ))}
                                />
                                <TextField
                                    style={{ width: "100%", marginBottom: "25px", marginTop: "10px" }}
                                    name="description"
                                    value={values.description}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    id="description"
                                    label="Description"
                                    multiline
                                    rowsMax={4}
                                />
                            </DialogContent>
                            <DialogActions>
                                <Button autoFocus onClick={handleClose} color="secondary">
                                    Cancel
                                </Button>
                                <Button autoFocus type="submit" color="primary" disabled={isSubmitting}>
                                    {buttonLabel}
                                </Button>
                            </DialogActions>
                        </form>
                    );
                }}
            </Formik>
        </Dialog>
    );
};

DIGform.propTypes = {
    onClose: PropTypes.func.isRequired,
    open: PropTypes.bool.isRequired,
};

export default DIGform;
