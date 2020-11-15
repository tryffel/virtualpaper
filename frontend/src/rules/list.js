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

import { List, Datagrid, TextField, EditButton, DateField, ChipField, BooleanField, Show, SimpleShowLayout} from 'react-admin';


const InfoPanel = props => (
    <Show
        {...props}
        title=" "
    >
        { props.record ?
            <SimpleShowLayout>
                <TextField source="filter"/>
                {props.record.action.date_fmt !== "" && <TextField label="Date format" source="action.date_fmt"/>}
                {props.record.action.date_separator !== "" && <TextField label="Date separator" source="action.date_separator" />}
                {props.record.action.description !== "" && <TextField label="Description" source="action.description" />}
                {props.record.action.metadata_key_id !== 0 && <TextField label="Metadata key" source="action.metadata_key_id" />}
                {props.record.action.metadata_key_value  !== 0 && <TextField label="Metadata value" source="action.metadata_value_id" />}
                }:
            </SimpleShowLayout>
            : null
        }
    </Show>
);

export const RuleList = (props) => (
    <List {...props}>
        <Datagrid expand={<InfoPanel />}>
            <TextField source="id" />
            <ChipField source="type" />
            <TextField label="Description" source="comment" />
            <BooleanField label="Enabled" source="active" />
            <DateField source="created_at" />
            <DateField source="updated_at" />
            <EditButton />
        </Datagrid>
    </List>
);




