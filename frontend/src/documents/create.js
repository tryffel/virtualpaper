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
import {Create, Error, FileField, FileInput, Loading, SimpleForm, useQueryWithStore} from "react-admin";
import {useEffect} from "react";



export const DocumentCreate = (props) => {
    const {data, loading, error} = useQueryWithStore({
        type: 'getOne',
        resource: 'filetypes',
        payload: {id: ""},
    });

    const [fileNames, setFileNames ] = React.useState('');
    const [mimeTypes, setMimeTypes ] = React.useState('');

    useEffect(() => {
        if (data) {
            setFileNames(data.names.join(', '));
            setMimeTypes(data.mimetypes.join(', '));
        }
    }, [data])

    if (loading) return <Loading />;
    if (error) return <Error error={error}/>;

    return (
    <Create{...props}>
        <SimpleForm title={"Upload new document"}>
            <span>Supported file types: {fileNames}</span>
            <FileInput accept={mimeTypes} multiple={false} label="File upload" source="file">
                <FileField source="src" title="title"/>
            </FileInput>
        </SimpleForm>
    </Create>
    )};

