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
    DateInput,
    Edit,
    SimpleForm,
    TextInput,
    DateField,
    TextField,
    ReferenceArrayInput,
    ReferenceInput,
    SelectArrayInput,
    useQueryWithStore,
    Loading,
    Error,
    SelectInput,
    ArrayInput,
    SimpleFormIterator,
    FormDataConsumer
} from "react-admin";

import MarkDownInputWithField from "ra-input-markdown";

import get from 'lodash/get';


export const DocumentEdit = (props) => {

    const transform = data => ({
        ...data,
        date: Date.parse(`${data.date}`),
    });

    return (
    <Edit {...props} transform={transform}>
        <SimpleForm>
            <TextField disabled label="Id" source="id" />
            <TextInput source="name" fullWidth />
            <MarkDownInputWithField source="description" />
            <DateInput source="date" />
            <ReferenceArrayInput source="tags" reference="tags" allowEmpty label={"Tags"}>
                <SelectArrayInput optionText="key" />
            </ReferenceArrayInput>
                <ArrayInput source="metadata" label={"Metadata"}>
                    <SimpleFormIterator margin="dense" defaultValue={ [{key_id: 0, key:"", value_id: 0, value:""}]}>
                        <ReferenceInput label="Key" source="key_id" reference="metadata/keys" fullWidth>
                            <SelectInput optionText="key" fullWidth/>
                        </ReferenceInput>
                        <FormDataConsumer>
                            {({getSource, scopedFormData}) =>
                                scopedFormData && scopedFormData.key_id ? (
                                    <MetadataValueInput
                                        source={getSource('value_id')}
                                        record={scopedFormData}
                                        label={"Value"}
                                        fullWidth
                                    />
                                ) : null
                            }

                        </FormDataConsumer>
                    </SimpleFormIterator>
                </ArrayInput>
            <DateField source="created_at" />
            <DateField source="updated_at" />
        </SimpleForm>
    </Edit>
    );
}

const MetadataValueInput = props => {
    let keyId = 0;
    if (props.record) {
        keyId = get(props.record, "key_id");
    }

    const {data, loading, error } = useQueryWithStore({
        type: 'getList',
        resource: 'metadata/values',
        payload: { target:"metadata/values", id: keyId!==0 ? keyId : -1,
            pagination: {page:1}, perPage: 200,
            sort:"id", order:"asc"}
    });

    if (loading) return <Loading />;
    if (error) return <Error error={error}/>;
    if (data) {
        return (
            <SelectInput {...props} choices={data} optionText="value" />

        )} else {
        return <Loading />;

    }
};
