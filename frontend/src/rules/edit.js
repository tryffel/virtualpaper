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

import * as React from "react";
import {
    Edit,
    SimpleForm,
    TextInput,
    RadioButtonGroupInput,
    BooleanInput,
    ReferenceInput,
    SelectInput, ArrayInput, SimpleFormIterator,
    useInput, Labeled, FormDataConsumer, useRecordContext,
    FormWithRedirect, SaveButton, DeleteButton
} from 'react-admin';

import { Typography, Box, Grid, Toolbar } from '@material-ui/core';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';

import MarkDownInputWithField from "ra-input-markdown";



export const RuleEdit = (props) => (
    <Edit {...props} title={"Edit process rule"}>
        <SimpleForm>
            <Typography variant="h5">Processing Rule</Typography>
            <BooleanInput label="Enabled" source="enabled"/>
            <TextInput source="name" fullWidth={true} />
            <MarkDownInputWithField source="description" fullWidth={true} />

            <RadioButtonGroupInput label="Match conditions" source="type" fullWidth={true} choices={[
                { id: 'match_all', name: 'Match all'},
                { id: 'match_any', name: 'Match any'},
            ]} />
            <Typography variant="h5">Rule Conditions</Typography>
            <ArrayInput source="conditions" label="">
                <ConditionEdit/>
            </ArrayInput>
            <Typography variant="h5">Rule Actions</Typography>
            <ArrayInput source="actions" label="">
                <ActionEdit/>
            </ArrayInput>
        </SimpleForm>
    </Edit>
);


const ConditionTypeInput = (props) => {
    const {
        input,
        meta: { touched, error }
    } = useInput(props);


    return (
        <SelectInput {...props} choices={[
            {id: "name_is", name:"Name is" },
            {id: "name_starts", name:" Name starts" },
            {id: "name_contains", name:" Name contains" },

            {id: "description_is", name:" Description is" },
            {id: "description_starts", name:" Description starts" },
            {id: "description_contains", name:" Description contains" },

            {id: "content_is", name:" Text content" },
            {id: "content_starts", name:" Text content" },
            {id: "content_contains", name:" Text content" },

            {id: "date_is", name:" Date is" },
            {id: "date_after", name:" Date is" },
            {id: "date_before", name:" Date is" },

            {id: "metadata_has_key", name:" Metadata contains" },
            {id: "metadata_has_key_value", name:" Metadata contains key-value" },
            {id: "metadata_count", name:" Metadata count equals" },
            {id: "metadata_count_less_than", name:" Metadata count less than" },
            {id: "metadata_count_more_than", name:" Metadata count more than" },
        ]}/> )
}


const ActionTypeInput = (props) => {
    const {
        input,
        meta: { touched, error }
    } = useInput(props);

    return (
        <SelectInput {...props} choices={[
            {id: "name_set", name:"Set name" },
            {id: "name_append", name:"Append name" },
            {id: "description_set", name:"Set description" },
            {id: "description_append", name:"Append description" },
            {id: "metadata_add", name:"Add metadata" },
            {id: "metadata_remove", name:"Remove metadata" },
            {id: "date_set", name:"Set date" },
        ]}
        />
    )
}


DeleteButton.propTypes = {};
const ConditionEdit = (props) => {
    //const record = useRecordContext(props);

    return (
        <SimpleFormIterator {...props} >
            <FormWithRedirect {...props} render={formProps => (
                <form>
                    <Box p="1em">
                        <Box display="flex">
                            <Box flex={1} mr="1em">
                                <Grid container display="flex" spacing={1} >
                                    <Grid item flex={1} ml="0.5em">
                                        <ConditionTypeInput label="Type" source="condition_type"/>
                                    </Grid>
                                    <Grid item flex={1} mr="0.5em">
                                        <BooleanInput label="Enabled" source="enabled"/>
                                    </Grid>
                                    <Grid item flex={1} ml="0.5em">
                                        <BooleanInput label="Inverted" source="inverted"/>
                                    </Grid>
                                    <Grid item flex={1} ml="0.5em">
                                        <BooleanInput label="Case insensitive" source="case_insensitive"/>
                                    </Grid>
                                    <Grid item flex={1} ml="0.5em">
                                        <BooleanInput label="Regex" source="is_regex"/>
                                    </Grid>
                                </Grid>
                                <Grid container display="flex" spacing={2}>
                                    <Grid item flex={1} ml="0.5em">
                                        <TextInput label="Filter" source="value" fullWidth resettable/>
                                    </Grid>
                                    <Grid item flex={1} ml="0.5em">
                                        <TextInput label="Date format" source="date_fmt" fullWidth resettable/>
                                    </Grid>
                                </Grid>

                            </Box>
                        </Box>
                    </Box>
                </form>
            )} />
        </SimpleFormIterator>
    )
}


const ActionEdit = (props) => {
    //const record = useRecordContext(props);

    return (
        <SimpleFormIterator {...props}  >
            <FormWithRedirect {...props} render={formProps => (
                <form>
                    <Box p="1em">
                        <Box display="flex">
                            <Box flex={1} mr="1em">
                                <Grid container display="flex" spacing={2}>
                                    <Grid item flex={1} ml="0.5em">
                                        <ActionTypeInput label="Type" source="action"/>
                                    </Grid>
                                    <Grid item flex={1} ml="0.5em">
                                        <BooleanInput label="Enabled" source="enabled"/>
                                    </Grid>
                                    <Grid item flex={1} mr="0.5em">
                                        <BooleanInput label="On condition" source="on_condition"/>
                                    </Grid>
                                </Grid>
                                <Grid container display="flex" spacing={2}>
                                    <Grid item flex={1} ml="0.5em">
                                        <TextInput label="Format" source="value"/>
                                    </Grid>
                                </Grid>

                            </Box>
                        </Box>
                    </Box>
                </form>
            )} />
        </SimpleFormIterator>
    )
}