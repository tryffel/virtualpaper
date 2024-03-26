/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
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
import { Button } from "react-admin";
import HelpIcon from "@mui/icons-material/Help";
import {
  Dialog,
  DialogTitle,
  DialogContentText,
  DialogContent,
  DialogActions,
  Typography,
} from "@mui/material";

export interface HelpButtonProps {
  label?: string;
  title: string;
  children?: JSX.Element | JSX.Element[] | never;
}

export const HelpButton = (props: HelpButtonProps) => {
  const { label, ...rest } = props;
  const [open, setOpen] = React.useState(false);
  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  return (
    <div>
      <Button
        label={label ? label : "Help"}
        size="small"
        alignIcon="left"
        onClick={handleClickOpen}
      >
        <HelpIcon />
      </Button>
      <HelpDialog open={open} onClose={handleClose} {...rest} />
    </div>
  );
};

interface Props extends HelpButtonProps {
  onClose: () => void;
  open: boolean;
}

const HelpDialog = (props: Props) => {
  const [scroll] = React.useState("paper");
  const { onClose, open, title } = props;
  const handleClose = () => {
    onClose();
  };

  return (
    <Dialog
      onClose={handleClose}
      aria-labelledby="simple-dialog-title"
      open={open}
    >
      <DialogTitle id="simple-dialog-title">Help: {title}</DialogTitle>
      <DialogContent dividers={scroll === "paper"}>
        <DialogContentText>{props.children}</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>
          <Typography>Close</Typography>
        </Button>
      </DialogActions>
    </Dialog>
  );
};
