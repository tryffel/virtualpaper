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
    SelectInput,
} from 'react-admin';

import { Typography } from '@material-ui/core';


export const RuleEdit = (props) => (
    <Edit {...props} title={"Edit process rule"}>
        <SimpleForm>
            <Typography variant="h5">Rule trigger</Typography>
            <TextInput label="description" source="comment" fullWidth={true} />
            <RadioButtonGroupInput source="type" fullWidth={true} choices={[
                { id: 'regex', name: 'Regular expression' },
                { id: 'exact', name: 'Match' },
            ]} />
            <TextInput label="Filter expression" source="filter" fullWidth={true} />
            <BooleanInput label="Enabled" source="active"/>
            <Typography variant="h5">Action</Typography>
            <TextInput label="Date format" source="action.date_fmt" fullWidth={true}/>
            <TextInput label="Date separator" source="action.date_separator" fullWidth={true}/>
            <TextInput label="Description" source="action.description" fullWidth={true}/>
            <ReferenceInput source="action.tag_id" reference="tags" allowEmpty label="Tag">
                <SelectInput optionText="key" />
            </ReferenceInput>
        </SimpleForm>
    </Edit>
);

