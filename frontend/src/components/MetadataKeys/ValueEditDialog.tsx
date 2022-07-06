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

import React from "react";
import {
  required,
  Button,
  SaveButton,
  TextInput,
  BooleanInput,
  useUpdate,
  useNotify,
  useRefresh,
  RadioButtonGroupInput,
  Form,
} from "react-admin";

import { Cancel } from "@mui/icons-material";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { useFormState } from "react-hook-form";

const MetadataValueUpdateDialog = (props: any) => {
  const [update, { data, isLoading, error }] = useUpdate("metadata/values", {
    id: props.record?.id,
    data: undefined,
    previousData: props.record,
  });
  const notify = useNotify();
  const refresh = useRefresh();

  const handleCloseClick = () => {
    props.setShowDialog(false);
  };

  const handleSubmit = async (values: any) => {
    update("metadata/values", {
      data: values,
      id: props.record.id,
      // @ts-ignore
      key_id: props.key_id,
      meta: { key_id: props.key_id },
    });
  };

  if (isLoading) {
    props.setShowDialog(false);
    refresh();
  }

  if (error) {
    console.info(error);
    // @ts-ignore
    notify(error.message, "error");
  }

  return (
    <>
      <Dialog
        fullWidth
        open={props.showDialog}
        onClose={handleCloseClick}
        aria-label="Update metadata value"
      >
        <DialogTitle>Update metadata</DialogTitle>
        <Form
          resource="metadata/value"
          onSubmit={handleSubmit}
          warnWhenUnsavedChanges={true}
          {...props}
        >
          <DialogContent>
            <TextInput source="value" validate={required()} fullWidth />
            <TextInput label="description" source="comment" fullWidth />
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
              fullWidth
            />
          </DialogContent>
          <DialogActions>
            <Button
              label="ra.action.cancel"
              onClick={handleCloseClick}
              // @ts-ignore
            >
              <Cancel />
            </Button>
            <SaveButton />
          </DialogActions>
        </Form>
      </Dialog>
    </>
  );
};

export default MetadataValueUpdateDialog;
