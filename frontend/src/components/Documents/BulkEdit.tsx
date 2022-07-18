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
import { useParams, useSearchParams } from "react-router-dom";
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
  TextInput,
  TextField,
  useStore,
} from "react-admin";
import {
  Typography,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import { ExpandMore } from "@mui/icons-material";
import { DocumentCard } from "./List";
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

  //const [metadataAdd, setMetadataAdd] = React.useState<Metadata[]>([]);
  //const [metadataRemove, setMetadataRemove] = React.useState<Metadata[]>([]);

  const emptyRecord = {
    documents: ids,
    add_metadata: {metadata: []},
    remove_metadata: {metadata: []},
  };

  if (isLoading) {
    return <Loading />;
  }

  return (
    <CreateBase record={emptyRecord} redirect="false" >
      <SimpleForm>
        <Accordion>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography variant="h5">Documents</Typography>
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
        <Accordion>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography variant="h5">Add metadata</Typography>
          </AccordionSummary>
          <AccordionDetails style={{ flexDirection: "column" }}>
            <ArrayInput source="add_metadata.metadata" label={"Add metadata"}>
              <SimpleFormIterator
                defaultValue={[{ key_id: 0, key: "", value_id: 0, value: "" }]}
                disableReordering={true}
              >
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
        <Accordion>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography variant="h5">Remove metadata</Typography>
          </AccordionSummary>
          <AccordionDetails style={{ flexDirection: "column" }}>
            <ArrayInput source="remove_metadata.metadata" label={"Add metadata"}>
              <SimpleFormIterator
                defaultValue={[{ key_id: 0, key: "", value_id: 0, value: "" }]}
                disableReordering={true}
              >
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
      </SimpleForm>
    </CreateBase>
  );
};

const AddMetadata = (props: any) => {
  const { metadataList, setMetadataList } = props;

  const defaultMetadata = {
    KeyId: 0,
    Key: "",
    ValueId: 0,
    Value: "",
  };

  const addDefault = () => {
    setMetadataList(metadataList.concat([defaultMetadata]));
  };

  console.log("to add: ", metadataList);

  return (
    <>
      <Button label="Add new" onClick={addDefault}></Button>

      {metadataList.map((metadata: any) => (
        <p>{metadata.Key}</p>
      ))}
    </>
  );
};

export default BulkEditDocuments;
