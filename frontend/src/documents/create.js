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
import {Create, FileField, FileInput, SimpleForm, TextInput} from "react-admin";


export const DocumentCreate = (props) => (
    <Create{...props}>
        <SimpleForm>
            <TextInput disabled source="id" label="id"/>
            <TextInput source="name" label="name" />
            <FileInput accept="application/pdf" multiple={false} label="File upload" source="file">
                <FileField source="src" title="title" />
            </FileInput>

        </SimpleForm>
    </Create>
);

