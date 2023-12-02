import * as React from "react";
import {
  ArrayInput,
  BooleanInput,
  Button,
  Form,
  HttpError,
  NumberInput,
  RecordContext,
  SaveButton,
  SimpleFormIterator,
  TextField,
  useNotify,
  useRecordContext,
  useUpdate,
} from "react-admin";
import ShareIcon from "@mui/icons-material/Share";
import MenuItem from "@mui/material/MenuItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Typography,
} from "@mui/material";
import { RequestIndexSelect } from "./RequestIndexing";

export const EditDocumentSharing = (props: { onClose: () => void }) => {
  const [step, setStep] = React.useState("fts");
  const record = useRecordContext();
  const notify = useNotify();
  const [open, setOpen] = React.useState(false);
  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    props.onClose();
  };

  return (
    <>
      <MenuItem onClick={handleClickOpen}>
        <ListItemIcon>
          <ShareIcon color={"primary"} />
        </ListItemIcon>
        <ListItemText>
          <Typography variant="body1" color={"primary"}>
            Edit sharing
          </Typography>
        </ListItemText>
      </MenuItem>
      {open && <EditDialogContent onClose={handleClose} open={open} />}
    </>
  );
};

const EditDialogContent = ({
  onClose,
  open,
}: {
  onClose: () => void;
  open: boolean;
}) => {
  const notify = useNotify();
  const record = useRecordContext();

  const [update, { isLoading }] = useUpdate(
    "document-user-sharing",
    undefined,
    {
      onSuccess: () => {
        notify("Updated");
        onClose();
      },
      onError: (err: HttpError) => {
        notify(`Error: ${err.message} (status: ${err.status})`, {
          type: "error",
        });
      },
    }
  );

  const handleUpdate = (data: any) => {
    console.log("saving", data);
    update("document-user-sharing", { id: record.id, data: data });
  };

  const formId = `edit-document-${record.id}-sharing`;

  return (
    <RecordContext.Provider
      value={{ users: record.shared_users, id: record.id }}
    >
      <Form warnWhenUnsavedChanges onSubmit={handleUpdate} id={formId}>
        <Dialog
          onClose={onClose}
          aria-labelledby="simple-dialog-title"
          open={open}
        >
          <DialogTitle id="simple-dialog-title">
            Share document with users
          </DialogTitle>
          <DialogContent>
            <DialogContentText>
              <ArrayInput
                source={"users"}
                defaultValue={{
                  user_id: 0,
                  permissions: { read: true, write: false, delete: false },
                }}
              >
                <SimpleFormIterator inline disableReordering>
                  <NumberInput source={"user_id"} label={"User Id"} />
                  <TextField source={"user_name"} label={"Username"} />
                  <BooleanInput
                    source={"permissions.read"}
                    label={"Can view"}
                  />
                  <BooleanInput
                    source={"permissions.write"}
                    label={"Can edit"}
                  />
                  <BooleanInput
                    source={"permissions.delete"}
                    label={"Can delete"}
                  />
                </SimpleFormIterator>
              </ArrayInput>
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <SaveButton form={formId} />
            <Button onClick={onClose} label={"Close"} />
          </DialogActions>
        </Dialog>
      </Form>
    </RecordContext.Provider>
  );
};
