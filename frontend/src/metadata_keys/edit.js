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
import {
    Datagrid,
    DateField,
    TextField,
    ReferenceManyField,
    Labeled,
    SimpleForm,
    Edit,
    BooleanField
} from "react-admin";

import { MarkdownField } from '../markdown'

import MetadataValueCreateButton from './valueCreate'
import MetadataValueUpdateDialog from "./valueEdit";
import {useState} from "react";


export const MetadataKeyEdit = (props) => {
    const [showUpdateDialog, setShowUpdateDialog] = useState(false);
    const [valueToUpdate, setValueToUpdate] = useState({id: 35});

    const onClickValue = (id, basePath, record) => {
        setValueToUpdate({record: record, key_id: props.id, id: record.id, basePath: "metadata/keys/" + props.id + "/values"});
        setShowUpdateDialog(true);

    }

    return (
    <Edit {...props} title={props.record ? props.record.key: 'Metadata key'}>
        <SimpleForm>
            <MetadataValueUpdateDialog
                showDialog={showUpdateDialog}
                setShowDialog={setShowUpdateDialog}
                {...props}
                basePath={valueToUpdate.basePath}
                resource="metadata/values"
                {...valueToUpdate}
            />

            <TextField source="key"/>
            <Labeled label="Description">
                <MarkdownField source="description"/>
            </Labeled>

            <ReferenceManyField  label="Values" reference={"metadata/values"} target={"key_id"}>
                <Datagrid rowClick={onClickValue}>
                    <TextField source="value"/>
                    <BooleanField label="Automatic matching" source="match_documents"/>
                    <TextField label="Match by" source="match_type"/>
                    <TextField label="Filter" source="match_filter"/>
                </Datagrid>
            </ReferenceManyField>

            <MetadataValueCreateButton />

            <DateField source="created_at" showTime={false}/>
        </SimpleForm>
    </Edit>
    );
};


