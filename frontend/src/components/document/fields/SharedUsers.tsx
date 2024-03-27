import get from "lodash/get";
import { useRecordContext } from "react-admin";
import { Chip, Paper } from "@mui/material";
import ListSubheader from "@mui/material/ListSubheader";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Collapse from "@mui/material/Collapse";
import ExpandLess from "@mui/icons-material/ExpandLess";
import ExpandMore from "@mui/icons-material/ExpandMore";
import React from "react";
import ListItem from "@mui/material/ListItem";
import Avatar from "@mui/material/Avatar";
import PersonIcon from "@mui/icons-material/Person";
import ListItemAvatar from "@mui/material/ListItemAvatar";
import PeopleIcon from "@mui/icons-material/People";
import { pink } from "@mui/material/colors";

export type Permissions = {
  read: boolean;
  write: boolean;
  delete: boolean;
};

export type SharedUser = {
  user_id: number;
  user_name: string;
  permissions: Permissions;
};

export const SharedUsersList = () => {
  const [open, setOpen] = React.useState(false);

  const record = useRecordContext();
  if (
    !get(record, "shared_users") ||
    get(record, "shared_users").length === 0
  ) {
    return null;
  }

  return (
    <Paper elevation={1}>
      <List subheader={<ListSubheader>Document sharing</ListSubheader>}>
        <ListItemButton onClick={() => setOpen(!open)}>
          <ListItemIcon>
            <PeopleIcon color={"primary"} />
          </ListItemIcon>
          <ListItemText
            primary={`${record.shared_users.length} ${record.shared_users.length > 1 ? "users" : "user"}`}
          />
          {open ? <ExpandLess /> : <ExpandMore />}
        </ListItemButton>
        <Collapse in={open} timeout={"auto"} unmountOnExit>
          <List>
            {record.shared_users.map((entry: SharedUser) => (
              <SharedUser entry={entry} />
            ))}
          </List>
        </Collapse>
      </List>
    </Paper>
  );
};

const SharedUser = ({ entry }: { entry: SharedUser }) => {
  return (
    <ListItem key={entry.user_id}>
      <ListItemAvatar>
        <Avatar sx={{ bgcolor: pink[500] }}>
          <PersonIcon />
        </Avatar>
      </ListItemAvatar>
      <ListItemText>{entry.user_name}</ListItemText>
      {entry.permissions.write && (
        <ListItemIcon>
          <Chip
            sx={{ margin: 1 }}
            label={"Can edit"}
            color={"warning"}
            variant={"outlined"}
          />
        </ListItemIcon>
      )}
    </ListItem>
  );
};
