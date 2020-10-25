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



import { Card, CardActions, CardContent, CardHeader, CardActionArea } from '@material-ui/core';

import { List, TextField, useListContext, DateField, EditButton, ShowButton, Filter, TextInput, RichTextField} from "react-admin";

import {ThumbnailSmall} from "./file";

import '../App.css';


const cardStyle = {
    width: 280,
    minHeight: 400,
    margin: '0.5em',
    display: 'inline-block',
    verticalAlign: 'top',
};

const DocumentFilter = (props) => (
    <Filter {...props}>
        <TextInput label="Search" source="q" alwaysOn />
    </Filter>
);


const DocumentGrid = () => {
    const { ids, data, basePath } = useListContext();

    return (
        <div style={{ margin: '1em' }}>
            {ids.map(id =>
                <Card key={id} style={cardStyle}>
                    <CardActionArea>
                        <CardHeader
                            title={<RichTextField record={data[id]} source="name"  style={{'.em': {'background-color':'#FFFF00'}}} />}
                            subheader={<DateField record={data[id]} source="created_at" />}
                        />
                        <CardContent>

                            <ThumbnailSmall component="img" url={data[id].preview_url} title="Img" />
                        </CardContent>
                    </CardActionArea>
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
    <List title="All documents" filters={<DocumentFilter />} {...props}>
        <DocumentGrid />
    </List>
);
