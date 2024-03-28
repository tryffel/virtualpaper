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
  useGetMany,
  Loading,
  Button,
  CreateBase,
  SimpleForm,
  useNotify,
  useRedirect,
  TopToolbar,
  DateInput,
  SaveButton,
} from "react-admin";
import {
  Typography,
  Box,
  Container,
  Paper,
} from "@mui/material";
import ClearIcon from "@mui/icons-material/Clear";
import { HelpButton } from "../../Help.tsx";
import { DocumentCard } from "@components/document/card";
import { LanguageSelectInput } from "@resources/Documents/Edit.tsx";
import { useSearchParams } from "react-router-dom";
import React from "react";
import {
  MetadataArrayInput,
} from "@components/document/edit/MetadataInput.tsx";

type BulkForm = {
  language: string;
  date: Date | null;
};

const BulkEditForm = ({ ids }: { ids: string[] }) => {
  const { data, isLoading } = useGetMany("documents", {
    ids,
  });
  const notify = useNotify();
  const redirect = useRedirect();

  const onSuccess = () => {
    notify(`Documents modified`);
    redirect("list", "documents");
  };

  const emptyRecord = {
    documents: ids,
    add_metadata: { metadata: [] },
    remove_metadata: { metadata: [] },
  };

  const cancel = () => {
    redirect("list", "documents");
  };

  if (isLoading) {
    return <Loading />;
  }

  const transform = (data: BulkForm) => ({
    ...data,
    date: Date.parse(`${data.date}`),
  });

  return (
    <CreateBase
      record={emptyRecord}
      redirect="false"
      mutationOptions={{ onSuccess }}
      transform={transform}
    >
      <Container>
        <Paper elevation={2}>
          <SimpleForm toolbar={<Toolbar cancel={cancel} />}>
            <Box
              sx={{
                display: "flex",
                flexDirection: "column",
                justifyItems: "space-between",
                alignItems: "flex-start",
                ml: 1,
                mr: 1,
                gap: "20px",
              }}
            >
              <Box
                sx={{ width: "100%", maxHeight: "50vh", overflow: "scroll" }}
              >
                <Typography variant="body1" color="text.secondary">
                  {ids ? "Editing " + ids.length + " documents" : null}
                </Typography>
                {data
                  ? data.map((document) => <DocumentCard record={document} />)
                  : null}
              </Box>
              <MetadataArrayInput
                source={"add_metadata.metadata"}
                label={"Add metadata"}
              />
              <MetadataArrayInput
                source={"remove_metadata.metadata"}
                label={"Remove metadata"}
              />
              <LanguageSelectInput source={"lang"} label={"Language"} />
              <DateInput source="date" />
            </Box>
          </SimpleForm>
        </Paper>
      </Container>
    </CreateBase>
  );
};

const BulkEditDocuments = () => {
  const [params] = useSearchParams();
  const ids = params.get("documents");
  const array = React.useMemo(() => {
    if (!ids) {
      console.error("document id-array parameter is empty");
      return [];
    }
    return JSON.parse(ids) as string[];
  }, [ids]);

  return <BulkEditForm ids={array} />;
};

const Toolbar = (props: any) => {
  const { cancel } = props;

  return (
    <TopToolbar
      sx={{
        display: "flex",
        flexDirection: "row",
        justifyItems: "space-between",
        alignItems: "center",
        ml: 1,
        mr: 1,
        gap: "10px",
      }}
    >
      <SaveButton />
      <BulkEditHelp />
      <Button
        size={"medium"}
        variant={"contained"}
        label="Cancel"
        startIcon={<ClearIcon />}
        onClick={cancel}
        sx={{ ml: "auto", mr: 0 }}
      />
    </TopToolbar>
  );
};

const BulkEditHelp = () => {
  return (
    <HelpButton title="Edit Multiple Documents">
      <p>
        With this form it is possible to edit multiple document simultaneously.
        This is particularly useful when there's multiple documents, maybe even
        defined with a filter, that need similar editing, such as removing or
        adding metadata.
      </p>

      <Typography variant="h6" color="textPrimary">
        Usage
      </Typography>
      <p>
        On top there's a list of documents that are being modified. Be sure to
        verify that the documents are indeed the ones that should be modified.
      </p>

      <ul>
        <li>Add metadata: adds one or more metadata key-values to documents</li>
        <li>
          Remove metadata: removes one or more metadata key-values from
          documents, if they have one.{" "}
        </li>
      </ul>
    </HelpButton>
  );
};

export default BulkEditDocuments;
