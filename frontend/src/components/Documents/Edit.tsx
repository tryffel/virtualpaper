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
  SelectInput,
  ArrayInput,
  SimpleFormIterator,
  FormDataConsumer,
  AutocompleteInput,
  useGetManyReference,
  Labeled,
  Toolbar,
  SaveButton,
  DeleteWithConfirmButton,
  Button,
  TopToolbar,
  ShowButton,
  useRecordContext,
} from "react-admin";

import { MarkdownInput } from "../Markdown";
import { Typography, Grid, Box, useMediaQuery } from "@mui/material";
import LinkIcon from "@mui/icons-material/Link";
import ArticleIcon from "@mui/icons-material/Article";
import get from "lodash/get";
import { IndexingStatusField } from "./IndexingStatus";
import { EmbedFile } from "./Thumbnail";
import { EditLinkedDocuments } from "./EditLinkedDocuments";

const EditToolBar = () => {
  return (
    <Toolbar>
      <SaveButton />
      <DeleteWithConfirmButton
        confirmTitle="Are you sure you want to move document to trash bin?"
        confirmContent="Document can be restored from the trash bin. It will be automatically deleted after 14 days."
      />
    </Toolbar>
  );
};

export const DocumentEdit = () => {
  const transform = (data: any) => ({
    ...data,
    date: Date.parse(`${data.date}`),
  });

  const [previewOpen, setPreviewOpen] = React.useState(false);
  const isMedium = useMediaQuery((theme: any) => theme.breakpoints.up("md"));

  return (
    <Edit
      transform={transform}
      title="Edit document"
      aside={
        <ToggledEmbedFile
          source="download_url"
          shown={previewOpen && isMedium}
        />
      }
      actions={
        <DocumentEditActions open={previewOpen} setOpen={setPreviewOpen} />
      }
    >
      <SimpleForm warnWhenUnsavedChanges toolbar={<EditToolBar />}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={10} lg={10}>
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
              <EditLinkedDocumentsButton />
            </Box>
            <Box display={{ xs: "block", sm: "block" }}>
              <ArrayInput source="metadata" label={"Metadata"}>
                <SimpleFormIterator inline disableReordering fullWidth>
                  <ReferenceInput
                    label="Key"
                    source="key_id"
                    reference="metadata/keys"
                    fullWidth
                    className="MuiBox"
                  >
                    <SelectInput optionText="key" data-testid="metadata-key" />
                  </ReferenceInput>

                  <FormDataConsumer>
                    {({ formData, scopedFormData, getSource }) =>
                      scopedFormData && scopedFormData.key_id ? (
                        <MetadataValueInput
                          source={getSource ? getSource("value_id") : ""}
                          record={scopedFormData}
                          label={"Value"}
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
      </SimpleForm>
    </Edit>
  );
};

export interface MetadataValueInputProps {
  source: string;
  record: any;
  label: string;
  fullWidth?: boolean;
}

export const MetadataValueInput = (props: MetadataValueInputProps) => {
  let keyId = 0;
  if (props.record) {
    // @ts-ignore
    keyId = get(props.record, "key_id");
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

const ToggledEmbedFile = (props: any) => {
  const { shown, source, filename } = props;
  if (!shown) return null;
  const isLg = useMediaQuery((theme: any) => theme.breakpoints.up("lg"));

  return (
    <Box sx={{ width: isLg ? "900px" : "400px" }}>
      <EmbedFile source={source} filename={filename} />
    </Box>
  );
};

const DocumentEditActions = (props: { open: any; setOpen: any }) => {
  const isMedium = useMediaQuery((theme: any) => theme.breakpoints.down("md"));
  const { open, setOpen } = props;
  const onClick = (e: any) => {
    setOpen(!open);
  };

  return (
    <TopToolbar>
      {!isMedium ? (
        <Button
          color="primary"
          label="Toggle preview"
          startIcon={<ArticleIcon />}
          onClick={onClick}
        ></Button>
      ) : null}
      <ShowButton />
    </TopToolbar>
  );
};

const EditLinkedDocumentsButton = () => {
  const record = useRecordContext();
  const [open, setOpen] = React.useState(false);

  return (
    <>
      <EditLinkedDocuments
        modalOpen={open}
        setModalOpen={setOpen}
        documentId={String(record.id)}
      />
      <Button onClick={() => setOpen(true)} label={"Linked documents"}>
        <LinkIcon />
      </Button>
    </>
  );
};
