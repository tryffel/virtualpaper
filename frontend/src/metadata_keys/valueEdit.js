/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2021  Tero Vierimaa
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
  FormWithRedirect,
  useRefresh,
  RadioButtonGroupInput,
} from "react-admin";
import IconCancel from "@material-ui/icons/Cancel";

import Dialog from "@material-ui/core/Dialog";
import DialogTitle from "@material-ui/core/DialogTitle";
import DialogContent from "@material-ui/core/DialogContent";
import DialogActions from "@material-ui/core/DialogActions";

const MetadataValueUpdateDialog = (props) => {
  const [update, { loading, loaded, error }] = useUpdate("metadata/values");
  const notify = useNotify();
  const refresh = useRefresh();

  const handleCloseClick = () => {
    props.setShowDialog(false);
  };

  const handleSubmit = async (values) => {
    update(
      { payload: { data: values, key_id: props.key_id } },
      props.record.id
    );
  };

  if (loaded) {
    props.setShowDialog(false);
    refresh();
  }

  if (error) {
    console.info(error);
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

        <FormWithRedirect
          resource="metadata/value"
          save={handleSubmit}
          warnWhenUnsavedChanges={true}
          {...props}
          render={({ handleSubmitWithRedirect, pristine, saving }) => (
            <>
              <DialogContent>
                <TextInput source="value" validate={required()} fullWidth />
                <TextInput label="description" source="comment" fullWidth />
                <BooleanInput
                  label="Automatic matching"
                  source="match_documents"
                />
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
                  disabled={loading}
                >
                  <IconCancel />
                </Button>
                <SaveButton
                  handleSubmitWithRedirect={handleSubmitWithRedirect}
                  pristine={pristine}
                  saving={saving}
                  disabled={loading}
                />
              </DialogActions>
            </>
          )}
        />
      </Dialog>
    </>
  );
};

export default MetadataValueUpdateDialog;
