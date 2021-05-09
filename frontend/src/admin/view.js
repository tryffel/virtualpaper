/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2021  Tero Vierimaa
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

import {Show, TextField, SimpleShowLayout} from "react-admin";


export const AdminView = ({staticContext, ...props}) => {
    return (

        <Show
            redirect={false}
            id="systeminfo"
            resource="admin"
            basePath="/admin"
            title="Administrating"
            {...props}
        >
            <SimpleShowLayout>
                <TextField source="name" />
                <TextField source="version" />
                <TextField source="commit" />
                <TextField source="uptime" />
                <TextField source="imagemagick_version" />
                <TextField source="tesseract_version" />
                <TextField source="poppler_installed" />
                <TextField source="go_version" />
            </SimpleShowLayout>
        </Show>

    );
}


export default AdminView;

