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
  useGetMany,
  Loading,
  Button,
  CreateBase,
  SimpleForm,
  ArrayInput,
  ReferenceInput,
  SimpleFormIterator,
  FormDataConsumer,
  SelectInput,
  useStore,
  useNotify,
  useRedirect,
  TopToolbar,
} from "react-admin";
import {
  Typography,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Box,
} from "@mui/material";
import { ExpandMore, Clear } from "@mui/icons-material";
import { HelpButton } from "../Help";
import { DocumentCard } from "./DocumentCard";
import { MetadataValueInput } from "./Edit";

interface Metadata {
  KeyId: number;
  Key: string;
  Valueid: number;
  Value: string;
}
interface body {
  documents: string[];
  addMetadata: Metadata[];
  removeMetadata: Metadata[];
}

const BulkEditDocuments = () => {
  const [documentIds, setStore] = useStore("bulk-edit-document-ids", []);
  // @ts-ignore
  const idList = documentIds;
  const ids = documentIds;
  console.log("ids to edit: ", idList);
  const { data, isLoading, error, refetch } = useGetMany("documents", {
    ids: idList,
  });
  const notify = useNotify();
  const redirect = useRedirect();

  const onSuccess = (data: any) => {
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

  return (
    <CreateBase
      record={emptyRecord}
      redirect="false"
      mutationOptions={{ onSuccess }}
    >
      <SimpleForm>
        <Toolbar cancel={cancel} />
        <Box width="100%">
          <Accordion>
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography variant="h5" sx={{ width: "33%" }}>
                Documents
              </Typography>
              <Typography variant="body1" color="text.secondary">
                {idList ? "Editing " + idList.length + " documents" : null}
              </Typography>
            </AccordionSummary>
            <AccordionDetails style={{ flexDirection: "column" }}>
              <Typography variant="body1">
                {data ? data.length : "0"} Documents to edit
              </Typography>
              {data
                ? data.map((document) => <DocumentCard record={document} />)
                : null}
            </AccordionDetails>
          </Accordion>
        </Box>
        <Box width="100%">
          <Accordion>
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography variant="h5" sx={{ width: "33%" }}>
                Add metadata
              </Typography>
            </AccordionSummary>
            <AccordionDetails style={{ flexDirection: "column" }}>
              <ArrayInput source="add_metadata.metadata" label={"Add metadata"}>
                <SimpleFormIterator disableReordering={true}>
                  <ReferenceInput
                    label="Key"
                    source="key_id"
                    reference="metadata/keys"
                    fullWidth
                    className="MuiBox"
                  >
                    <SelectInput
                      optionText="key"
                      fullWidth
                      data-testid="metadata-key"
                    />
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
            </AccordionDetails>
          </Accordion>
        </Box>
        <Box width="100%">
          <Accordion>
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography variant="h5">Remove metadata</Typography>
            </AccordionSummary>
            <AccordionDetails style={{ flexDirection: "column" }}>
              <ArrayInput
                source="remove_metadata.metadata"
                label={"Add metadata"}
              >
                <SimpleFormIterator disableReordering={true}>
                  <ReferenceInput
                    label="Key"
                    source="key_id"
                    reference="metadata/keys"
                    fullWidth
                    className="MuiBox"
                  >
                    <SelectInput
                      optionText="key"
                      fullWidth
                      data-testid="metadata-key"
                    />
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
            </AccordionDetails>
          </Accordion>
        </Box>
      </SimpleForm>
    </CreateBase>
  );
};

const Toolbar = (props: any) => {
  const { cancel } = props;

  return (
    <TopToolbar>
      <BulkEditHelp />
      <Button label="Cancel" startIcon={<Clear />} onClick={cancel} />
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
