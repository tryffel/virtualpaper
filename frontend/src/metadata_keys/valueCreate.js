/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */


import React, { useState } from 'react';
import {
    required,
    Button,
    SaveButton,
    TextInput,
    BooleanInput,
    useCreate,
    useNotify,
    FormWithRedirect,
    useRefresh, RadioButtonGroupInput, SimpleForm
} from 'react-admin';
import IconContentAdd from '@material-ui/icons/Add';
import IconCancel from '@material-ui/icons/Cancel';

import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';


function MetadataValueCreateButton({ onChange, record }) {
    const [showDialog, setShowDialog] = useState(false);
    const [create, { loading }] = useCreate('metadata/values', {value:""});
    const notify = useNotify();
    const refresh = useRefresh();

    const handleClick = () => {
        setShowDialog(true);
    };

    const handleCloseClick = () => {
        setShowDialog(false);
    };

    const handleSubmit = async values => {
        create(
            { payload: { data: values, key_id: record.id} },
            {
                onSuccess: () => {
                    setShowDialog(false);
                    refresh();
                },
                onFailure: ({ error }) => {
                    notify(error.message, 'error');
                }
            }
        );
    };

    return (
        <>
            <Button onClick={handleClick} label="ra.action.create">
                <IconContentAdd />
            </Button>
            <Dialog
                fullWidth
                open={showDialog}
                onClose={handleCloseClick}
                aria-label="Create new metadata value"
            >
                <DialogTitle>Add new metadata</DialogTitle>

                <FormWithRedirect
                    resource="metadata/keys"
                    save={handleSubmit}
                    render={({
                                 handleSubmitWithRedirect,
                                 pristine,
                                 saving
                             }) => (
                        <>
                            <DialogContent>
                                <TextInput source="value" validate={required()} fullWidth />
                                <TextInput label="description" source="comment" fullWidth />
                                <BooleanInput label="Automatic matching" source="match_documents"/>
                                <RadioButtonGroupInput source="match_type" fullWidth={true} choices={[
                                    { id: 'regex', name: 'Regular expression' },
                                    { id: 'exact', name: 'Match' },
                                ]} />
                                <TextInput label="Filter expression" source="match_filter" fullWidth={true} />
                            </DialogContent>
                            <DialogActions>
                                <Button
                                    label="ra.action.cancel"
                                    onClick={handleCloseClick}
                                    disabled={loading}
                                >
                                    <IconCancel />
                                </Button>
                                <SaveButton
                                    handleSubmitWithRedirect={
                                        handleSubmitWithRedirect
                                    }
                                    pristine={pristine}
                                    saving={saving}
                                    disabled={loading}
                                />
                            </DialogActions>
                        </>
                    )}
                />
            </Dialog>
        </>
    );
}

export default MetadataValueCreateButton;
