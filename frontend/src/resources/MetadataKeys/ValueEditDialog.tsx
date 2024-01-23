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
  useDelete,
  Confirm,
  RaRecord,
} from "react-admin";

import { Cancel } from "@mui/icons-material";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import { Link } from "react-router-dom";
import { EscapeWhitespace } from "../../components/util";

export interface MetadataValueUpdateDialogProps {
  showDialog: boolean;
  setShowDialog: (show: boolean) => void;
  basePath: string;
  resource: string;
  record?: RaRecord;
  keyId: number;
}

const MetadataValueUpdateDialog = (props: MetadataValueUpdateDialogProps) => {
  // temporary fix to break recursive loop inside Form component since React Admin 4.
  if (!props.showDialog) {
    return null;
  }

  const { showDialog, setShowDialog, record, keyId } = props;

  const [update, { error, isSuccess: updateSuccess }] = useUpdate(
    "metadata/values",
    {
      id: record?.id,
      data: {},
      previousData: record,
    },
  );

  const [deleteOne, { error: deleteError }] = useDelete("metadata/values", {
    id: record?.id,
    previousData: record,
  });

  const notify = useNotify();
  const refresh = useRefresh();
  const [confirmOpen, setConfirmOpen] = React.useState(false);

  const handleCloseClick = () => {
    setShowDialog(false);
  };

  const handleSubmit = (values: any) => {
    update("metadata/values", {
      data: values,
      id: record?.id,
      // @ts-ignore
      key_id: keyId,
      meta: { key_id: keyId },
    });
  };

  const onCancel = () => {
    setConfirmOpen(false);
  };

  const onConfirm = () => {
    deleteOne("metadata/values", {
      id: record?.id,
      // @ts-ignore
      key_id: keyId,
      meta: { key_id: keyId },
    });
    setConfirmOpen(false);
    handleCloseClick();
    notify("Metadata value deleted");
  };

  const handleDelete = async (values: any) => {
    setConfirmOpen(true);
  };

  if (updateSuccess) {
    setShowDialog(false);
    refresh();
  }

  if (error) {
    console.error(error);
    // @ts-ignore
    notify(error.message, "error");
  }

  if (deleteError) {
    console.error(deleteError);
    // @ts-ignore
    notify(deleteError.message, "error");
  }

  const linkDocsLabel = `Show documents (${
    props.record ? props.record.documents_count : ""
  })`;

  let to: any = {
    pathname: "/documents",
  };

  if (props.record) {
    to = {
      pathname: "/documents",
      search: `filter=${JSON.stringify({
        q:
          EscapeWhitespace(props.record.key) +
          ":" +
          EscapeWhitespace(props.record.value),
      })}`,
    };
  }

  return (
    <>
      <Confirm
        isOpen={confirmOpen}
        title={"Confirm deleting metadata value "}
        content={"This action cannot be undone"}
        onConfirm={onConfirm}
        onClose={onCancel}
      />

      <Dialog
        fullWidth
        open={showDialog}
        onClose={handleCloseClick}
        aria-label="Update metadata value"
      >
        <DialogTitle>Update metadata</DialogTitle>
        <Form onSubmit={handleSubmit} warnWhenUnsavedChanges={true} {...props}>
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
              label={linkDocsLabel}
              disabled={
                props.record ? props.record.documents_count === 0 : false
              }
              component={Link}
              to={to}
            />
            <Button
              label="ra.action.cancel"
              onClick={handleCloseClick}
              // @ts-ignore
            >
              <Cancel />
            </Button>
            <Button
              label="Delete"
              startIcon={<DeleteIcon />}
              // @ts-ignore
              sx={{ color: "secondary" }}
              onClick={handleDelete}
            />
            <SaveButton />
          </DialogActions>
        </Form>
      </Dialog>
    </>
  );
};

export default MetadataValueUpdateDialog;
