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
  AutocompleteInput,
  Button,
  DateInput,
  DeleteWithConfirmButton,
  Edit,
  Labeled,
  SaveButton,
  ShowButton,
  SimpleForm,
  TextInput,
  Toolbar,
  TopToolbar,
  useRecordContext,
} from "react-admin";

import { MarkdownInput } from "@components/markdown";
import { Box, Grid, Typography, useMediaQuery } from "@mui/material";
import LinkIcon from "@mui/icons-material/Link";
import ArticleIcon from "@mui/icons-material/Article";
import { IndexingStatusField } from "@components/document/fields/IndexingStatus.tsx";
import { EmbedFile } from "@components/document/fields/Thumbnail.tsx";
import { EditLinkedDocuments } from "@components/document/edit/EditLinkedDocuments.tsx";
import { languages } from "@/languages.ts";
import { DocumentIdField } from "@components/document/fields/DocumentId.tsx";
import { TimestampField } from "@components/primitives/TimestampField.tsx";
import { FavoriteDocumentInput } from "@components/document/edit/Favorite.tsx";
import { MetadataArrayInput } from "@components/document/edit/MetadataInput.tsx";

const EditToolBar = () => {
  return (
    <Toolbar>
      <SaveButton />
      <DeleteWithConfirmButton
        confirmTitle="Are you sure you want to move document to trash bin?"
        confirmContent="Document can be restored from the trash bin. It will be automatically deleted after 14 days."
        style={{ marginLeft: "auto" }}
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
      redirect={"show"}
    >
      <SimpleForm warnWhenUnsavedChanges toolbar={<EditToolBar />}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={10} lg={10}>
            <Box
              sx={{
                display: "flex",
                flexDirection: "row",
                gap: "10px",
                alignItems: "space-between",
              }}
            >
              <Box>
                <Typography variant="h6">Basic Info</Typography>
                <DocumentIdField />
              </Box>
              <Box sx={{ ml: "auto", mr: "10px" }}>
                <Labeled label={"Favorite"}>
                  <FavoriteDocumentInput source={"favorite"} />
                </Labeled>
              </Box>
              <Box sx={{ ml: "auto", mr: 0 }}>
                <IndexingStatusField source="status" showLabel />
              </Box>
            </Box>
          </Grid>
          <Grid item xs={12}>
            <TextInput source="name" fullWidth variant={"standard"} />
          </Grid>
          <Grid item xs={6} sm={6}>
            <DateInput source="date" fullWidth variant={"standard"} />
          </Grid>
          <Grid item xs={6} sm={6}>
            <LanguageSelectInput source={"lang"} label={"Language"} />
          </Grid>
          <Grid item xs={12} md={6}>
            <MarkdownInput source="description" label="Description" />
          </Grid>
          <Grid item xs={12}>
            <EditLinkedDocumentsButton />
          </Grid>
          <Grid item xs={12}>
            <MetadataArrayInput source={"metadata"} />
          </Grid>
          <Grid item xs={12}>
            <TimestampField />
          </Grid>
        </Grid>
      </SimpleForm>
    </Edit>
  );
};

export const LanguageSelectInput = (props: any) => {
  const choices = Object.keys(languages).map((key) => {
    return {
      id: key,
      name: languages[key as keyof typeof languages],
    };
  });

  return (
    <AutocompleteInput {...props} choices={choices} variant={"standard"} />
  );
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
  const onClick = () => {
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
