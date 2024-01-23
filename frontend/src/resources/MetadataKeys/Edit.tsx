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
} from "react-admin";

import { MarkdownInput } from "../../components/Markdown";
import { useMediaQuery } from "@mui/material";

import MetadataValueCreateButton from "./ValueCreate";
import MetadataValueUpdateDialog from "./ValueEditDialog";
import { useState } from "react";
import get from "lodash/get";
import { IconColorSelect, IconSelect } from "./IconSelect";

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

  return (
    <Edit
      title={<EditTitle />}
      transform={(data: any) => ({
        ...data,
        style: JSON.stringify(data.style),
      })}
    >
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
        <IconSelect source={"icon"} displayIcon={true} />
        <IconColorSelect />

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
