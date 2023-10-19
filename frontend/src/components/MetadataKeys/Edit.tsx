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
  Datagrid,
  DateField,
  TextField,
  ReferenceManyField,
  Labeled,
  SimpleForm,
  Edit,
  BooleanField,
  NumberField,
  useEditController,
  TextInput,
  useRecordContext,
  FunctionField,
  FormDataConsumer,
} from "react-admin";

import { MarkdownInput } from "../Markdown";
import { Typography, useMediaQuery } from "@mui/material";

import MetadataValueCreateButton from "./ValueCreate";
import MetadataValueUpdateDialog from "./ValueEditDialog";
import { useState } from "react";
import get from "lodash/get";
import { IconByName, iconExists } from "../icons";

export const MetadataKeyEdit = () => {
  const { record } = useEditController();
  const [keyId, setKeyId] = useState(0);

  const [showUpdateDialog, setShowUpdateDialog] = useState(false);
  const [valueToUpdate, setValueToUpdate] = useState({ id: 0, keyId: -1 });

  // @ts-ignore
  const onClickValue = (id, resource, record) => {
    setValueToUpdate({
      // @ts-ignore
      record: record,
      keyId: keyId,
      id: record.id,
      basePath: "metadata/keys/" + keyId + "/values",
    });
    setShowUpdateDialog(true);
  };
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  if (record && keyId == 0) {
    setKeyId(record.id);
  }

  const renderIcon = (record: any) => () => {
    console.log("record", record);
    if (!iconExists(record.icon)) {
      return <Typography color={"error"}>Icon does not exist</Typography>;
    } else {
      return <IconByName name={record.icon} />;
    }
  };

  const validateIcon = (value: any) => {
    if (iconExists(value)) {
      return undefined;
    }
    return "Icon is invalid. Must be Material-ui icon. Leave empty to disable";
  };

  const validateJson = (value: any) => {
    try {
      JSON.parse(value);
      return undefined;
    } catch (e) {
      return "invalid json";
    }
  };

  return (
    <Edit title={<EditTitle />}>
      <SimpleForm>
        <MetadataValueUpdateDialog
          showDialog={showUpdateDialog}
          setShowDialog={setShowUpdateDialog}
          // @ts-ignore
          basePath={valueToUpdate.basePath}
          resource="metadata/values"
          {...valueToUpdate}
        />
        <TextInput source="key" id="key-name" label="metadata key name" />
        <Labeled label={"Description"}>
          <MarkdownInput source="comment" />
        </Labeled>
        <TextInput
          source="icon"
          id="icon"
          label="Icon name (Material-ui)"
          validate={validateIcon}
        />
        <FormDataConsumer>
          {({ formData }) => {
            return <FunctionField render={renderIcon(formData)} />;
          }}
        </FormDataConsumer>
        <TextInput
          source="style"
          id="style"
          label="Style (json)"
          validate={validateJson}
        />
        <ReferenceManyField
          label="Values"
          reference={"metadata/values"}
          target={"key_id"}
          perPage={500}
          sortBy="Name"
          sortByOrder="ASC"
        >
          <Datagrid
            // @ts-ignore
            rowClick={onClickValue}
            bulkActionButtons={false}
          >
            <TextField source="value" />
            <BooleanField label="Automatic matching" source="match_documents" />
            {!isSmall ? (
              <TextField label="Match by" source="match_type" />
            ) : null}
            {!isSmall ? (
              <TextField label="Filter" source="match_filter" />
            ) : null}
            <NumberField source="documents_count" label={"Total documents"} />
          </Datagrid>
        </ReferenceManyField>

        <MetadataValueCreateButton record={record} />
        <Labeled label="Created at">
          <DateField source="created_at" showTime={false} />
        </Labeled>
      </SimpleForm>
    </Edit>
  );
};

const EditTitle = () => {
  const record = useRecordContext();
  const name = get(record, "key") ?? "";

  return <span>Metadata key {name}</span>;
};
