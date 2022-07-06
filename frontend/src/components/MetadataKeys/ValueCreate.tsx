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

import React, { useState } from "react";
import {
  required,
  Button,
  SaveButton,
  TextInput,
  BooleanInput,
  useCreate,
  useNotify,
  Form,
  useRefresh,
  RadioButtonGroupInput,
  useRecordContext,
} from "react-admin";

import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";

import { Create, Cancel } from "@mui/icons-material";

const MetadataValueCreateButton = (record: any) => {
  const [showDialog, setShowDialog] = useState(false);
  const [create, { data, isLoading, error }] = useCreate("metadata/values", {});

  const [waitingResponse, setWaitingResponse] = useState(false);

  const notify = useNotify();
  const refresh = useRefresh();

  const handleClick = () => {
    setShowDialog(true);
  };

  const handleCloseClick = () => {
    setShowDialog(false);
  };

  const handleSubmit = async (values: any) => {
    setWaitingResponse(true);
    create("metadata/values", { data: values, meta: { key_id: record.id } });
  };

  if (data && waitingResponse) {
    refresh();
    setShowDialog(false);
    setWaitingResponse(false);
  }

  if (error) {
    // @ts-ignore
    notify("Failed to create metadata key value: ", error.message);
  }

  return (
    <>
      <Button onClick={handleClick} label="ra.action.create">
        <Create />
      </Button>
      <Dialog
        fullWidth
        open={showDialog}
        onClose={handleCloseClick}
        aria-label="Create new metadata value"
      >
        <DialogTitle>Add new metadata</DialogTitle>

        <Form resource="metadata/keys" onSubmit={handleSubmit}>
          {" "}
          <DialogContent>
            <TextInput source="value" validate={required()} fullWidth />
            <BooleanInput label="Automatic matching" source="match_documents" />
            <RadioButtonGroupInput
              source="match_type"
              fullWidth={true}
              isRequired={true}
              defaultValue={"exact"}
              choices={[
                { id: "exact", name: "Exact match" },
                { id: "regex", name: "Regular expression" },
              ]}
            />
            <TextInput
              label="Filter expression"
              source="match_filter"
              fullWidth={true}
            />
          </DialogContent>
          <DialogActions>
            <Button label="ra.action.cancel" onClick={handleCloseClick}>
              <Cancel />
            </Button>
            <SaveButton />
          </DialogActions>
        </Form>
      </Dialog>
    </>
  );
};

export default MetadataValueCreateButton;
