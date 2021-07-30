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

import * as React from "react"

import get from 'lodash/get';
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    DateField,
    BooleanField,
    useRecordContext
} from 'react-admin';

import { Chip, Typography } from '@material-ui/core';


export const RuleList = (props) => (
    <List {...props}>
        <Datagrid>
            <TextField source="name" />
            <TextField source="id" />
            <TextField label="Description" source="description" />
            <BooleanField label="Enabled" source="enabled" />
            <RuleModeField source="mode" />
            <ChildCounterField source="conditions" />
            <ChildCounterField source="actions" />
            <DateField source="created_at" />
            <DateField source="updated_at" />
            <EditButton />
        </Datagrid>
    </List>
);

const RuleModeField = (props) => {
    const {source } = props;
    const record = useRecordContext(props);
    const value = get(record, source);

    return <Chip label={value === "match_all" ? "Match all": "Match any"}/>
}

const ChildCounterField = (props) => {
    const {source } = props;
    const record = useRecordContext(props);
    const value = get(record, source);

    return record ?
        <Typography component="span" variant="body2">
            {value ? value.length : ""}
        </Typography> : null
}



