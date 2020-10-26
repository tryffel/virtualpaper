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

import { List, TextField, Datagrid, Filter, TextInput, RichTextField, ChipField} from "react-admin";


const TagFilter = (props) => (
    <Filter {...props}>
        <TextInput label="Search" source="q" alwaysOn />
    </Filter>
);



export const TagList = (props) => (
    <List title="All tags" filters={<TagFilter/>} {...props}>
        <Datagrid rowClick="edit">
            <ChipField source="key" label={"Name"}/>
            <RichTextField source="comment" />
            <TextField source="document_count" label={"Documents"} />

        </Datagrid>
    </List>
);

