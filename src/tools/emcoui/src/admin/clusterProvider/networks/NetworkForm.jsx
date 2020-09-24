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
import React from 'react';
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
import { TextField, Select, InputLabel } from '@material-ui/core';
import * as Yup from "yup";
import { Formik } from 'formik';

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
        name: Yup.string().required("required"),
        description: Yup.string(),
        spec: Yup.object().typeError("Invalid network spec").required("Network spec is required")
    })

const NetworkForm = (props) => {
    const { onClose, item, open, onSubmit } = props;
    const buttonLabel = item ? "OK" : "Create"
    const title = item ? "Edit Network" : "Add Network"
    const handleClose = () => {
        onClose();
    };
    let initialValues = item ? { name: item.metadata.name, description: item.metadata.description } : { name: "", description: "", spec: undefined, type: "networks" }

    return (
        <Dialog maxWidth={"xs"} onClose={handleClose} aria-labelledby="customized-dialog-title" open={open} disableBackdropClick>
            <DialogTitle id="simple-dialog-title">{title}</DialogTitle>
            <Formik
                initialValues={initialValues}
                onSubmit={async values => {
                    onSubmit(values);
                }}
                validationSchema={schema}
            >
                {props => {
                    const {
                        values,
                        touched,
                        errors,
                        isSubmitting,
                        handleChange,
                        handleBlur,
                        handleSubmit
                    } = props;
                    return (
                        <form encType="multipart/form-data" noValidate onSubmit={handleSubmit}>
                            <DialogContent dividers>
                                <InputLabel htmlFor="age-native-simple">Type</InputLabel>
                                <Select
                                    native
                                    value={values.type}
                                    id="type"
                                    name='type'
                                    onChange={handleChange}
                                    onBlur={handleBlur}

                                >
                                    <option key="0" >provider-networks</option>
                                    <option key="1" >networks</option>
                                </Select>
                                <TextField
                                    style={{ width: "100%", marginBottom: "10px" }}
                                    id="name"
                                    label="Network name"
                                    type="text"
                                    value={values.name}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    helperText={(errors.name && touched.name && (
                                        (errors.name)
                                    ))}
                                    required
                                    error={errors.name && touched.name}
                                />
                                <TextField
                                    style={{ width: "100%", marginBottom: "10px" }}
                                    id="spec"
                                    label="Network Specs"
                                    type="text"
                                    value={values.spec}
                                    onChange={handleChange}
                                    onBlur={handleBlur}
                                    required
                                    multiline
                                    rows={5}
                                    variant="outlined"
                                    error={errors.spec && touched.spec}
                                    helperText={(errors.spec && touched.spec && (
                                        (errors["spec"])
                                    ))}
                                />
                                <TextField
                                    style={{ width: "100%", marginBottom: "25px" }}
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

NetworkForm.propTypes = {
    onClose: PropTypes.func.isRequired,
    open: PropTypes.bool.isRequired,
    item: PropTypes.object
};

export default NetworkForm;
