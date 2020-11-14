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

import React from 'react';
import { Edit, TextInput, SimpleForm, DateField, TextField } from 'react-admin';

export const ProfileEdit = ({ staticContext, ...props }) => {

    return (
        <Edit
            redirect={false}
            id="user"
            resource="preferences"
            basePath="/preferences"
            title="Profile"
            {...props}
        >
            <SimpleForm>
                <TextField source="user_name" label={"Username"}/>
                <TextInput source="email" />
                <DateField source="created_at" label={"Joined at"}/>
                <DateField  source="updated_at" label={"Last updated"}/>
                </SimpleForm>
        </Edit>
    );
};

