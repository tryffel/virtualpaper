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
import { Admin, Resource } from 'react-admin';

import { dataProvider } from "./dataProvider";
import authProvider from './authProvider';
import { DocumentList, DocumentShow, DocumentEdit, DocumentCreate } from './document';
import { JobList} from "./job";


const App = () => (
    <Admin dataProvider={dataProvider} authProvider={authProvider}>
        <Resource name="documents" list={DocumentList} show={DocumentShow} edit={DocumentEdit} create={DocumentCreate}/>
        <Resource name="jobs" list={JobList} />
    </Admin>
    );

export default App;

