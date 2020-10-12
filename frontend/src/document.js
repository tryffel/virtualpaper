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
import {Datagrid, List, TextField, Show, SimpleShowLayout, TextInput, Edit, SimpleForm } from "react-admin";


export const DocumentList = (props) => (
    <List {...props}>
        <Datagrid rowClick="show" >
            <TextField source="id" />
            <TextField source="Name" />
            <TextField source="Filename" />
            <TextField source="CreatedAt" />
            <TextField source="UpdatedAt" />
        </Datagrid>
    </List>
);

export const DocumentShow = (props) => (
    <Show {...props}>
        <SimpleShowLayout>
            <TextField source="id" />
            <TextField source="Name" />
            <TextField label="File" source="Filename" />
            <TextField source="Content" />
            <TextField source="CreatedAt" />
            <TextField source="UpdatedAt" />
        </SimpleShowLayout>
    </Show>
);

export const DocumentEdit = (props) => (
    <Edit {...props}>
        <SimpleForm>
            <TextInput disabled label="Id" source="id" />
            <TextInput source="title" />
            <TextInput multiline source="teaser" />
        </SimpleForm>
    </Edit>
);