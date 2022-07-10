/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
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
  Loading,
  Error,
  SelectInput,
  ArrayInput,
  SimpleFormIterator,
  FormDataConsumer,
  AutocompleteInput,
  useRecordContext,
  useGetList,
  useGetManyReference,
  Labeled,
} from "react-admin";

import { MarkdownInput } from "../Markdown";
import { Typography, Grid, Box } from "@mui/material";
import get from "lodash/get";
import "./Edit.css";
import { IndexingStatusField } from "./IndexingStatus";

export const DocumentEdit = () => {
  const transform = (data: any) => ({
    ...data,
    date: Date.parse(`${data.date}`),
  });

  return (
    <Edit transform={transform} title="Edit document">
      <SimpleForm redirect="show" warnWhenUnsavedChanges>
        <div>
          <Grid container spacing={2}>
            <Grid item xs={12} md={8} lg={12}>
              <Typography variant="h6">Basic Info</Typography>
              <Labeled label="Document id">
                <TextField label="Id" source="id" id="document-id" />
              </Labeled>
              <Box display={{ xs: "block", sm: "flex" }}>
                <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                  <TextInput source="name" fullWidth />
                </Box>
                <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
                  <DateInput source="date" />
                </Box>
                <IndexingStatusField source="status" />
              </Box>
              <Box display={{ xs: "block", sm: "flex" }}>
                <MarkdownInput source="description" label="Description" />
              </Box>
              <Box display={{ xs: "block", sm: "bloxk" }}>
                <ReferenceArrayInput
                  source="tags"
                  reference="tags"
                  allowEmpty
                  label={"Tags"}
                >
                  <SelectArrayInput optionText="key" />
                </ReferenceArrayInput>
              </Box>
              <Box display={{ xs: "block", sm: "block" }}>
                <ArrayInput source="metadata" label={"Metadata"}>
                  <SimpleFormIterator
                    defaultValue={[
                      { key_id: 0, key: "", value_id: 0, value: "" },
                    ]}
                    disableReordering={true}
                  >
                    <ReferenceInput
                      label="Key"
                      source="key_id"
                      reference="metadata/keys"
                      fullWidth
                      className="MuiBox"
                    >
                      <SelectInput optionText="key" fullWidth />
                    </ReferenceInput>

                    <FormDataConsumer>
                      {({ formData, scopedFormData, getSource }) =>
                        scopedFormData && scopedFormData.key_id ? (
                          <MetadataValueInput
                            source={getSource ? getSource("value_id") : ""}
                            record={scopedFormData}
                            label={"Value"}
                            fullWidth
                          />
                        ) : null
                      }
                    </FormDataConsumer>
                  </SimpleFormIterator>
                </ArrayInput>
              </Box>
              <Box display={{ xs: "block", sm: "block" }}>
                <Labeled label="Created at">
                  <DateField source="created_at" />
                </Labeled>
                <Labeled label="Updated at">
                  <DateField source="updated_at" />
                </Labeled>
              </Box>
            </Grid>
          </Grid>
        </div>
      </SimpleForm>
    </Edit>
  );
};

export interface MetadataValueInputProps {
  source: string;
  record: any;
  label: string;
  fullWidth: boolean;
}

const MetadataValueInput = (props: MetadataValueInputProps) => {
  let keyId = 0;
  if (props.record) {
    // @ts-ignore
    keyId = get(props.record, "key_id");
    console.log("key ", keyId);
  }
  const { data, isLoading, error } = useGetManyReference("metadata/values", {
    target: "id",
    id: keyId !== 0 ? keyId : -1,
    pagination: { page: 1, perPage: 500 },
    sort: {
      field: "value",
      order: "ASC",
    },
  });

  if (!props.record) {
    return null;
  }

  if (isLoading) return <Loading />;
  if (error) return <Typography>Error {error.message}</Typography>;
  if (data) {
    return (
      <AutocompleteInput
        {...props}
        choices={data}
        optionText="value"
        className="MuiBox"
      />
    );
  } else {
    return <Loading />;
  }
};
