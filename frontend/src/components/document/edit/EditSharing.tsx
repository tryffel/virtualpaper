import * as React from "react";
import {
  ArrayInput,
  BooleanInput,
  Button,
  Form,
  HttpError,
  RecordContext,
  ReferenceInput,
  SaveButton,
  SelectInput,
  SimpleFormIterator,
  useNotify,
  useRecordContext,
  useRefresh,
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
  Tooltip,
  Typography,
} from "@mui/material";

export const EditDocumentSharing = (props: { onClose: () => void }) => {
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
  const refresh = useRefresh();
  const notify = useNotify();
  const record = useRecordContext();

  const [update] = useUpdate("document-user-sharing", undefined, {
    onSuccess: () => {
      notify("Updated");
      refresh();
      onClose();
    },
    onError: (err: HttpError) => {
      notify(`Error: ${err.message} (status: ${err.status})`, {
        type: "error",
      });
    },
  });

  const handleUpdate = (data: any) => {
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
                label={""}
                defaultValue={{
                  user_id: 0,
                  permissions: { read: true, write: false, delete: false },
                }}
              >
                <SimpleFormIterator inline disableReordering>
                  <ReferenceInput source={"user_id"} reference={"users"}>
                    <SelectInput
                      source={"user_id"}
                      label={"User"}
                      optionText={"name"}
                      required
                    />
                  </ReferenceInput>
                  <Tooltip
                    title={"Does user have permission to edit the document"}
                  >
                    <BooleanInput
                      source={"permissions.write"}
                      label={"Can edit"}
                    />
                  </Tooltip>
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
