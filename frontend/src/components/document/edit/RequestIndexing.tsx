import * as React from "react";
import { Button, useNotify, useRecordContext } from "react-admin";
import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormControl,
  InputLabel,
  Select,
  Typography,
} from "@mui/material";
import { requestDocumentProcessing } from "@api/dataProvider.ts";
import AutoFixNormalIcon from "@mui/icons-material/AutoFixNormal";
import MenuItem from "@mui/material/MenuItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";

export const RequestIndexingModal = (props: { onClose: () => void }) => {
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

  const handleExecute = () => {
    if (record) {
      // @ts-ignore
      requestDocumentProcessing(record.id);
      notify(`Processing scheduled`, { type: "success" });
      setOpen(false);
    }
  };

  return (
    <>
      <MenuItem onClick={handleClickOpen}>
        <ListItemIcon>
          <AutoFixNormalIcon color={"primary"} />
        </ListItemIcon>
        <ListItemText>
          <Typography variant="body1" color={"primary"}>
            Process
          </Typography>
        </ListItemText>
      </MenuItem>

      <Dialog
        onClose={handleClose}
        aria-labelledby="simple-dialog-title"
        open={open}
      >
        <DialogTitle id="simple-dialog-title">
          Request document processing
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            <Typography variant={"body2"}>
              Are you sure you want to request processing?
            </Typography>
            <RequestIndexSelect step={step} setStep={setStep} />
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>
            <Typography>Close</Typography>
          </Button>
          <Button onClick={handleExecute} color="secondary" variant="contained">
            <Typography>Execute</Typography>
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export const RequestIndexSelect = (props: {
  step: string;
  setStep: (step: string) => void;
  enabledAll?: boolean;
}) => {
  const { step, setStep } = props;
  const handleChangeStep = (event: any) => {
    setStep(event.target.value as string);
  };
  const disabled = !props.enabledAll;

  return (
    <FormControl fullWidth>
      <InputLabel id="step">Starting step</InputLabel>
      <Select
        labelId="step"
        id="step"
        value={step}
        label="Age"
        onChange={handleChangeStep}
      >
        <MenuItem value={"thumbnail"} disabled={disabled}>
          Thumbnail
        </MenuItem>
        <MenuItem value={"content"} disabled={disabled}>
          Extract
        </MenuItem>
        <MenuItem value={"detect-language"} disabled={disabled}>
          Detect language
        </MenuItem>
        <MenuItem value={"rules"}>Rules</MenuItem>
        <MenuItem value={"fts"} disabled={disabled}>
          Index
        </MenuItem>
      </Select>
    </FormControl>
  );
};
