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



import { Card, CardActions, CardContent, CardHeader, CardActionArea, Box, useMediaQuery, Typography } from '@material-ui/core';
import CircularProgress from '@material-ui/core/CircularProgress';

import {
    List,
    useListContext,
    DateField,
    EditButton,
    ShowButton,
    Filter,
    TextInput,
    RichTextField,
    Pagination,
    TopToolbar,
    SortButton,
    CreateButton,
    ExportButton,
} from "react-admin";

import { ThumbnailSmall } from "./file";
import { FilterSidebar } from './filter';

import '../App.css';
import {useState} from "react";


const cardStyle = {
    width: 280,
    minHeight: 400,
    margin: '0.5em',
    display: 'inline-block',
    verticalAlign: 'top',
};

const DocumentPagination = props => <Pagination rowsPerPageOptions={[10, 25, 50, 100]} {...props} />;


const DocumentFilter = (props) => {

    return (
        <Filter {...props} >
            <TextInput label="Search" source="q" alwaysOn resettable={true}/>
            <TextInput label="Metadata (k.v)" source="metadata" alwaysOn/>
        </Filter>

    );
}


const DocumentSearchFilter = (props) => {
    return (
        <Filter {...props} >
                <TextInput label="Search" source="q" alwaysOn resettable={true}/>
            <TextInput label="Metadata (k.v)" source="metadata"/>
        </Filter>
    );
}


const DocumentGrid = () => {
    const { ids, data, basePath } = useListContext();

    return (
        ids ?
        <Box style={{ margin: '1em' }}>
            {ids.map(id =>
                <Card key={id} style={cardStyle}>
                    <CardActionArea>
                        <CardHeader
                            title={<RichTextField record={data[id]} source="name"  style={{'.em': {'background-color':'#FFFF00'}}} />}
                            subheader={<DateField record={data[id]} source="date" />}
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
        </Box>
            : null
    );
};

export const DocumentList = (props) => {
    const isSmall = useMediaQuery(theme => theme.breakpoints.down('sm'));
    if (isSmall) return <SmallDocumentList {...props}/>
    else return <LargeDocumentList {...props}/>;
}

const SmallDocumentList = (props) => {
    return (
        <List
            title="All documents"
            pagination={<DocumentPagination/>}
            filters={<DocumentSearchFilter/>}
            {...props}
        ><DocumentGrid/>
    </List>
    )
}


const LargeDocumentList = (props) => {
    return (
        <List
            title="All documents"
            pagination={<DocumentPagination/>}
            aside={<FilterSidebar/>}
            filters={<DocumentFilter/>}
            actions={<DocumentListActions />}
            sort={{field: 'date', order: 'DESC'}}
            {...props}
        ><DocumentGrid/>
        </List>
    )
}


export const IndexingStatusField = (props) => {

    const [value, setValue] = useState('ready')
    const [label, setLabel] = useState('Ready')
    if (props.record && props.source) {
        const status = props.record[props.source]
        if (status === 'indexing' && value !== status) {
            setValue(status);
            setLabel('Indexing document')

        } else if (status === 'pending' && value !== status) {
            setValue(status);
            setLabel('Indexing pending')
        } else if (status === 'ready' && value !== status) {
            setValue(status);
            setLabel("Ready")
        }
    }

    return (
        (value === 'ready') ?
            null :
            <Box>
                <CircularProgress variant="indeterminate" size={25} {...props} />
                <Typography variant="caption" component="div" color="textSecondary">{label}</Typography>
            </Box>
    );
}


const DocumentListActions = () => (
    <TopToolbar>
        <SortButton label="Sort" fields={['date', 'name', 'updated_at', 'created_at']} />
        <CreateButton basePath="/documents" />
        <ExportButton />
    </TopToolbar>
);
