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

import React, { useState } from "react";
import {ArrayField, Datagrid, DateField, Show, Tab, TabbedShowLayout, TextField, ChipField, SingleFieldList, Labeled, TopToolbar, EditButton, useNotify, useQuery
} from "react-admin";
import Button from '@material-ui/core/Button';
import RepeatIcon from '@material-ui/icons/Repeat';

import { ThumbnailField, EmbedFile} from "./file";
import { MarkdownField } from '../markdown'
import { IndexingStatusField } from "./list";
import { requestDocumentProcessing } from "../dataProvider";


export const DocumentShow = (props) => {
    const [enableFormatting ,setState]=useState(true);

    const toggleFormatting = () => {
        if (enableFormatting) {
            setState(false);
        } else {
            setState(true);
        }
    }

    return (

        <Show {...props} title="Document" actions={<DocumentShowActions/>} >
            <TabbedShowLayout>
                <Tab label="general">
                    <TextField source="name" label="" style={{fontSize:'2em'}}  />
                    <DateField source="date" showTime={false} label=""/>
                    <IndexingStatusField source="status" label=""/>
                    <ThumbnailField source="preview_url"/>
                    <Labeled label="Description">
                        <MarkdownField source="description"/>
                    </Labeled>
                    <ArrayField source="tags">
                        <SingleFieldList>
                            <ChipField source="key"/>
                        </SingleFieldList>
                    </ArrayField>

                    <ArrayField source="metadata">
                        <Datagrid>
                            <TextField source="key"/>
                            <TextField source="value"/>
                        </Datagrid>
                    </ArrayField>
                    <DateField source="created_at" label="Uploaded" showTime={false}/>
                    <DateField source="updated_at" label="Last updated" showTime={true}/>
                </Tab>
                <Tab label="Content">
                    <Button label={enableFormatting?"Enable formatting":"Disable formatting"} onClick={toggleFormatting}/>
                    {enableFormatting?<TextField source="content" label="Raw parsed text content"/>:
                    <MarkdownField source="content" label="Raw parsed text content"/>}
                </Tab>
                <Tab label="preview">
                    <EmbedFile source="download_url"/>
                </Tab>

            </TabbedShowLayout>
        </Show>
    );
}

const DocumentShowActions = ({ basePath, data, resource }) => {
    const requestProcessing = () => {
        if (data) {
            requestDocumentProcessing(data.id)
        }
    }

    return (
    <TopToolbar>
        <EditButton basePath={basePath} record={data}/>
        <Button color="primary" startIcon={<RepeatIcon/>} onClick={requestProcessing} >Request re-processing</Button>
    </TopToolbar>
    );
}

