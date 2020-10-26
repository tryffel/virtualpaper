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
import {DateInput, Edit, SimpleForm, TextInput, DateField, TextField} from "react-admin";


export const DocumentEdit = (props) => {

    const transform = data => ({
        ...data,
        date: Date.parse(`${data.date}`),
    });

    return (
    <Edit {...props} transform={transform}>
        <SimpleForm>
            <TextField disabled label="Id" source="id" />
            <TextInput source="name" />
            <TextInput source="description" />
            <DateInput source="date" />
            <DateField source="created_at" />
            <DateField source="updated_at" />
        </SimpleForm>
    </Edit>
    );
}
