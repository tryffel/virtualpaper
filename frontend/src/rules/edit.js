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
} from 'react-admin';

import { Typography } from '@material-ui/core';
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
            <ArrayInput source="actions">
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
        <Labeled label="Condition type">
            <Select {...input} >
                <MenuItem value="name_is">Name is</MenuItem>
                <MenuItem value="name_starts"> Name starts</MenuItem>
                <MenuItem value="name_contains"> Name contains</MenuItem>

                <MenuItem value="description_is"> Description is</MenuItem>
                <MenuItem value="description_starts"> Description starts</MenuItem>
                <MenuItem value="description_contains"> Description contains</MenuItem>

                <MenuItem value="content_is"> Text content is</MenuItem>
                <MenuItem value="content_starts"> Text content starts</MenuItem>
                <MenuItem value="content_contains"> Text content contains</MenuItem>

                <MenuItem value="date_is"> Date is</MenuItem>
                <MenuItem value="date_after"> Date is after</MenuItem>
                <MenuItem value="date_before"> Date is before</MenuItem>

                <MenuItem value="metadata_has_key"> Metadata contains key</MenuItem>
                <MenuItem value="metadata_has_key_value"> Metadata contains key-value</MenuItem>
                <MenuItem value="metadata_count"> Metadata count equals</MenuItem>
                <MenuItem value="metadata_count_less_than"> Metadata count less than</MenuItem>
                <MenuItem value="metadata_count_more_than"> Metadata count more than</MenuItem>
            </Select>
        </Labeled>
    )
}


const ActionTypeInput = (props) => {
    const {
        input,
        meta: { touched, error }
    } = useInput(props);

    return (
        <Labeled label="Action type">
            <Select {...input} >
                <MenuItem value="name_set">Set name</MenuItem>
                <MenuItem value="name_append">Append name</MenuItem>
                <MenuItem value="description_set">Set description</MenuItem>
                <MenuItem value="description_append">Append description</MenuItem>
                <MenuItem value="metadata_add">Add metadata</MenuItem>
                <MenuItem value="metadata_remove">Remove metadata</MenuItem>
                <MenuItem value="date_set">Set date</MenuItem>
            </Select>
        </Labeled>
    )
}


const ConditionEdit = (props) => {
    //const record = useRecordContext(props);

    return (
        <SimpleFormIterator {...props}>
            <BooleanInput label="Enabled" source="enabled"/>
            <BooleanInput label="Inverted" source="inverted"/>
            <BooleanInput label="Case insensitive" source="case_insensitive"/>
            <ConditionTypeInput label="Type" source="condition_type"/>
            <BooleanInput label="Regex" source="is_regex"/>
            <TextInput label="Filter" source="value"/>
            <TextInput label="Date format" source="date_fmt"/>

        </SimpleFormIterator>
    )
}


const ActionEdit = (props) => {
    //const record = useRecordContext(props);

    return (
        <SimpleFormIterator {...props}>
            <BooleanInput label="Enabled" source="enabled"/>
            <BooleanInput label="On condition" source="on_condition"/>
            <ActionTypeInput label="Type" source="action"/>
            <TextInput label="Format" source="value"/>
        </SimpleFormIterator>
    )
}