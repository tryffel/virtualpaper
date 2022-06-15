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
    Edit,
    SimpleForm,
    TextInput,
    RadioButtonGroupInput,
    BooleanInput,
    ReferenceInput,
    SelectInput, ArrayInput, SimpleFormIterator,
    FormDataConsumer,
    useQueryWithStore, Loading, Error, AutocompleteInput
} from 'react-admin';
import get from 'lodash/get';

import { Typography, Grid} from '@material-ui/core';
import MarkDownInputWithField from "ra-input-markdown";


export const RuleEdit = (props) => (
    <Edit {...props} title={"Edit process rule"}>
        <SimpleForm>
            <Typography variant="h5">Processing Rule edit</Typography>
            <BooleanInput label="Enabled" source="enabled"/>
            <TextInput source="name" fullWidth={true} />
            <MarkDownInputWithField source="description" fullWidth={true} />

            <RadioButtonGroupInput label="Match conditions" source="mode" fullWidth={true} choices={[
                { id: 'match_all', name: 'Match all'},
                { id: 'match_any', name: 'Match any'},
            ]} />
            <Typography variant="h5">Rule Conditions</Typography>
            <ArrayInput source="conditions" label="">
                <ConditionEdit/>
            </ArrayInput>
            <Typography variant="h5">Rule Actions</Typography>
            <ArrayInput source="actions" label="">
                <ActionEdit/>
            </ArrayInput>
        </SimpleForm>
    </Edit>
);


export const ConditionTypeInput = (props) => {
    return (
        <SelectInput {...props} onChange={props.onChange} choices={[
            {id: "name_is", name:"Name is" },
            {id: "name_starts", name:" Name starts" },
            {id: "name_contains", name:" Name contains" },

            {id: "description_is", name:" Description is" },
            {id: "description_starts", name:" Description starts" },
            {id: "description_contains", name:" Description contains" },

            {id: "content_is", name:" Text content matches" },
            {id: "content_starts", name:" Text content starts with" },
            {id: "content_contains", name:" Text content contains" },

            {id: "date_is", name:" Date is" },
            {id: "date_after", name:" Date is after" },
            {id: "date_before", name:" Date is before" },

            {id: "metadata_has_key", name:" Metadata contains" },
            {id: "metadata_has_key_value", name:" Metadata contains key-value" },
            {id: "metadata_count", name:" Metadata count equals" },
            {id: "metadata_count_less_than", name:" Metadata count less than" },
            {id: "metadata_count_more_than", name:" Metadata count more than" },
        ]} /> )
}


export const ActionTypeInput = (props) => {
    return (
        <SelectInput {...props} choices={[
            {id: "name_set", name:"Set name" },
            {id: "name_append", name:"Append name" },
            {id: "description_set", name:"Set description" },
            {id: "description_append", name:"Append description" },
            {id: "metadata_add", name:"Add metadata" },
            {id: "metadata_remove", name:"Remove metadata" },
            {id: "date_set", name:"Set date" },
        ]}
        />
    )
}

export const ConditionEdit = (props) => {

    return (
        props.record ?
            <SimpleFormIterator source={props.source} defaultValue={{enabled: false,
                condition_type: "content_contains",
                value: "empty"}}>
                <FormDataConsumer>
                    {({getSource, scopedFormData}) => {
                        return (
                            <Grid container spacing={1}>
                                <Grid container spacing={1}>
                                    <Grid item xs={6} md={3} lg={2}>
                                        <BooleanInput label="Enabled" source={getSource("enabled")} record={scopedFormData} initialValue={true}/>
                                    </Grid>
                                    <Grid item xs={6} md={3} lg={2}>
                                        <BooleanInput label="Case insensitive" source={getSource("case_insensitive")} record={scopedFormData}/>
                                    </Grid>
                                    <Grid item xs={6} md={3} lg={2}>
                                        <BooleanInput label="Inverted" source={getSource("inverted")} record={scopedFormData}/>
                                    </Grid>
                                    <Grid item xs={6} md={3} lg={2}>
                                        <BooleanInput label="Regex" source={getSource("is_regex")} record={scopedFormData}/>
                                    </Grid>
                                </Grid>
                                <Grid container spacing={1}>
                                    <Grid item xs={8} md={5} lg={3}>
                                        <ConditionTypeInput label="Type" source={getSource("condition_type")} record={scopedFormData}/>
                                    </Grid>
                                    {scopedFormData && scopedFormData.condition_type && !scopedFormData.condition_type.startsWith('date') &&
                                    !scopedFormData.condition_type.startsWith('metadata_has_key') ?
                                        <Grid item xs={8} md={5} lg={3}>
                                            <TextInput label="Filter" source={getSource("value")}
                                                       record={scopedFormData} fullWidth resettable/>
                                        </Grid> : null
                                    }
                                    {scopedFormData && scopedFormData.condition_type && scopedFormData.condition_type.startsWith('date') ?
                                        <Grid item xs={8} md={4} lg={3}>
                                            <TextInput label="Date format" source={getSource("date_fmt")}
                                                       record={{scopedFormData}} fullWidth resettable/>
                                        </Grid>: null
                                    }
                                    {scopedFormData && scopedFormData.condition_type && scopedFormData.condition_type.startsWith('metadata_has_key') ?
                                        <Grid item xs={8} md={4} lg={3}>
                                            <ReferenceInput label="Key" source={getSource("metadata.key_id")}
                                                            record={scopedFormData} reference="metadata/keys" fullWidth>
                                                <SelectInput optionText="key" fullWidth/>
                                            </ReferenceInput>
                                        </Grid>: null
                                    }
                                    {scopedFormData && scopedFormData.condition_type && scopedFormData.condition_type === 'metadata_has_key_value' ?
                                        <Grid item xs={8} md={4} lg={3}>
                                            <MetadataValueInput
                                                source={getSource('metadata.value_id')}
                                                keySource={'metadata.key_id'}
                                                record={scopedFormData}
                                                label={"Value"}
                                                fullWidth
                                            />
                                        </Grid>: null
                                    }

                                </Grid>
                            </Grid>
                        )
                    }}
                </FormDataConsumer>
            </SimpleFormIterator>: null
    )
}


export const ActionEdit = (props) => {
    return (
        props.record ?
            <SimpleFormIterator source={props.source} defaultValue={{enabled: false,
                action: "name_append",
                value: ""}}>
                <FormDataConsumer>
                    {({getSource, scopedFormData}) => {
                        return (
                                <Grid container spacing={2}>
                                    <Grid container spacing={2}>
                                        <Grid container display="flex" spacing={2}>
                                            <Grid item flex={1} ml="0.5em">
                                                <ActionTypeInput
                                                    label="Type"
                                                    source={getSource("action")}
                                                    record={scopedFormData}/>
                                            </Grid>
                                            <Grid item flex={1} ml="0.5em">
                                                <BooleanInput
                                                    label="Enabled"
                                                    source={getSource("enabled")}
                                                    record={scopedFormData}/>
                                            </Grid>
                                            <Grid item flex={1} mr="0.5em">
                                                <BooleanInput
                                                    label="On condition"
                                                    source={getSource("on_condition")}
                                                    record={scopedFormData}/>
                                            </Grid>
                                        </Grid>
                                        {scopedFormData && scopedFormData.action && !scopedFormData.action.startsWith('metadata') ?
                                            <Grid container display="flex" spacing={2}>
                                                <Grid item flex={1} ml="0.5em">
                                                    <TextInput label="Value" source="value"/>
                                                </Grid>
                                            </Grid>: null
                                        }
                                        {scopedFormData && scopedFormData.action && scopedFormData.action.startsWith('metadata') ?
                                            <Grid container display="flex" spacing={2}>
                                                <Grid item xs={8} md={4} lg={3}>
                                                    <ReferenceInput label="Key" source={getSource("metadata.key_id")}
                                                                    record={scopedFormData} reference="metadata/keys" fullWidth>
                                                        <SelectInput optionText="key" fullWidth/>
                                                    </ReferenceInput>
                                                </Grid>
                                                <Grid item xs={8} md={4} lg={3}>
                                                    <MetadataValueInput
                                                        source={getSource('metadata.value_id')}
                                                        keySource={'metadata.key_id'}
                                                        record={scopedFormData}
                                                        label={"Value"}
                                                        fullWidth
                                                    />
                                                </Grid>
                                            </Grid>: null
                                        }
                                    </Grid>
                                </Grid>
                        )}}
                </FormDataConsumer>
                            </SimpleFormIterator>: null
                    )
}

const MetadataValueInput = (props) => {
    let keyId = 0;
    if (props.record) {
        keyId = get(props.record, props.keySource);
    }

    const {data, loading, error } = useQueryWithStore({
        type: 'getList',
        resource: 'metadata/values',
        payload: { target:"metadata/values", id: keyId!==0 ? keyId : -1,
            pagination: {page:1, perPage: 500},
            sort: {
                field: "value",
                order: "ASC",
            },
        }
    });

    if (!props.record) {
        return null;
    }

    if (loading) return <Loading />;
    if (error) return <Error error={error}/>;
    if (data) {
        return (
            <AutocompleteInput{...props} choices={data} optionText="value" />

        )} else {
        return <Loading />;
    }
};
