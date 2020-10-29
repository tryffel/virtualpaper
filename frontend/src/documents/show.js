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
import {ArrayField, Datagrid, DateField, FileField, RichTextField, Show, Tab, TabbedShowLayout, TextField, ChipField, SingleFieldList, Labeled
} from "react-admin";

import { ThumbnailField, EmbedFile} from "./file";
import { MarkdownField } from '../markdown'


export const DocumentShow = (props) => (
    <Show {...props}>
        <TabbedShowLayout>
            <Tab label="general">
                <ThumbnailField source="preview_url" />
                <TextField source="id" />
                <TextField source="name" />

                <Labeled label="Description"  >
                    <MarkdownField source="description" />
                </Labeled>
                <DateField source="date" showTime={false} />
                <TextField source="pretty_size" label="Size"/>
                <TextField source="status" />



                <ArrayField source="tags">
                    <SingleFieldList>
                        <ChipField source="key" />
                    </SingleFieldList>
                </ArrayField>

                <ArrayField source="metadata">
                    <Datagrid>
                        <TextField source="key" />
                        <TextField source="value" />
                    </Datagrid>
                </ArrayField>
                <FileField source="download_url" label="Download document" title={"filename"} />
                <DateField source="created_at" showTime={true} />
                <DateField source="updated_at" showTime={true}/>
            </Tab>
            <Tab label="Plain text">
                <RichTextField source="content" />
            </Tab>
            <Tab label="preview">
                <EmbedFile source="download_url" />
            </Tab>

        </TabbedShowLayout>
    </Show>
);
