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
import { useState } from 'react';
import { List, TextField, Show, TextInput, Edit, SimpleForm,
    useListContext, DateField, EditButton, ShowButton, FileInput, Create, RichTextField, TabbedShowLayout,
    Tab, FileField, DateInput, DateTimeInput} from "react-admin";
import { Card, CardActions, CardContent, CardHeader } from '@material-ui/core';
import PropTypes from 'prop-types';
import get from 'lodash/get';


const cardStyle = {
    width: 400,
    minHeight: 400,
    margin: '0.5em',
    display: 'inline-block',
    verticalAlign: 'top'
};

function downloadFile(url) {
    const  token  = localStorage.getItem('auth');
    return fetch(url, {
        method: "GET",
        headers: {"Authorization": `Bearer ${token}`}
    })
}


function ThumbnailField ({ source, record })
{
    const url = get(record, source);
    const [imgData, setImage] = useState(() => {
        downloadFile(url)
            .then(response => {
                response.arrayBuffer().then(function (buffer) {
                    const data = window.URL.createObjectURL(new Blob([buffer]));
                    setImage(data);
                });
            })
            .catch( response => {
                    console.log(response);
                }
            );
        return "";
    });

    return (
        <div>
            <img src={imgData}/>
        </div>
    );
}

ThumbnailField.propTypes = {
    label: PropTypes.string,
    record: PropTypes.object,
    source: PropTypes.string.isRequired,
};


function ThumbnailSmall ({ url })
{
    const [imgData, setImage] = useState(() => {
        downloadFile(url)
            .then(response => {
                response.arrayBuffer().then(function (buffer) {
                    const data = window.URL.createObjectURL(new Blob([buffer]));
                    setImage(data);
                });
            })
            .catch( response => {
                    console.log(response);
                }
            );
        return "";
    });

    return (
        <div>
            <img src={imgData}/>
        </div>
    );
}

function EmbedFile({ source, record })
{
    const url = get(record, source);
    console.log("Download file: ", url);
    const [imgData, setImage] = useState(() => {
        downloadFile(url)
            .then(response => {
                response.arrayBuffer().then(function (buffer) {
                    const data = window.URL.createObjectURL(new Blob([buffer]));
                    setImage(data);
                });
            })
            .catch( response => {
                    console.log(response);
                }
            );
        return "";
    });

    return (
        <div style={{display:'block', width:'100%'}}>
            <iframe style={{width: '100%', display: 'fill', border: 'none', height:'40em'}} title="Preview" src={imgData}/>
        </div>
    );
}

EmbedFile.propTypes = {
    label: PropTypes.string,
    record: PropTypes.object,
    source: PropTypes.string.isRequired,
};


const DocumentGrid = () => {
    const { ids, data, basePath } = useListContext();

    return (
        <div style={{ margin: '1em' }}>
            {ids.map(id =>
                <Card key={id} style={cardStyle}>
                    <CardHeader
                        title={<TextField record={data[id]} source="name" />}
                        subheader={<DateField record={data[id]} source="created_at" />}
                    />
                    <CardContent>
                        <ThumbnailSmall url={data[id].preview_url} title="Img" />
                    </CardContent>
                    <CardActions style={{ textAlign: 'right' }}>
                        <ShowButton resource="documents" basePath={basePath} record={data[id]} />
                        <EditButton resource="documents" basePath={basePath} record={data[id]} />
                    </CardActions>
                </Card>
            )}
        </div>
    );
};

export const DocumentList = (props) => (
    <List title="All documents" {...props}>
        <DocumentGrid />
    </List>
);


export const DocumentShow = (props) => (
    <Show {...props}>
        <TabbedShowLayout>
            <Tab label="general">
                <ThumbnailField source="preview_url" />
                <TextField source="id" />
                <TextField source="name" />
                <TextField source="pretty_size" label="Size"/>
                <TextField source="status" />
                <DateField source="date" showTime={false} />
                <DateField source="created_at" showTime={true} />
                <DateField source="updated_at" showTime={true}/>
                <FileField source="download_url" label="Download document" title={"filename"} />
            </Tab>
            <Tab label="content">
                <RichTextField source="content" />
            </Tab>
            <Tab label="preview">
                <EmbedFile source="download_url" />
            </Tab>

        </TabbedShowLayout>
    </Show>
);

export const DocumentEdit = (props) => (
    <Edit {...props}>
        <SimpleForm>
            <TextInput disabled label="Id" source="id" />
            <TextInput source="name" />
            <DateInput source="date" />
            <DateTimeInput source="created_at" disabled />
            <DateTimeInput source="updated_at" disabled />
        </SimpleForm>
    </Edit>
);

export const DocumentCreate = (props) => (
    <Create{...props}>
        <SimpleForm>
            <TextInput source="id" label="id"/>
            <TextInput source="name" label="name" />
            <FileInput accept="application/pdf" multiple={false} label="File upload" >
                <FileField source="file" title="title" />
            </FileInput>

    </SimpleForm>
    </Create>
);

