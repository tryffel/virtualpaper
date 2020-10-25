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
import {ArrayField, Datagrid, DateField, FileField, RichTextField, Show, Tab, TabbedShowLayout, TextField
} from "react-admin";

import { ThumbnailField, EmbedFile} from "./file";


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

                <ArrayField source="metadata">
                    <Datagrid>
                        <TextField source="key" />
                        <TextField source="value" />
                    </Datagrid>
                </ArrayField>
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
