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
import { useQuery } from "react-query";
import {
  Loading,
  Error,
  useDataProvider,
  useAuthProvider,
  UserMenu as RaUserMenu,
  Logout,
} from "react-admin";
import { Link } from "react-router-dom";
import MenuItem from "@mui/material/MenuItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import SettingsIcon from "@mui/icons-material/Settings";
import AdminPanelSettingsIcon from "@mui/icons-material/AdminPanelSettings";

const UserMenuPreferences = React.forwardRef((props: any, ref: any) => {
  const dataProvider = useDataProvider();

  const { data, isLoading, error } = useQuery(["preferences", "user"], () =>
    dataProvider.getOne("preferences", { id: "user" })
  );

  if (isLoading) return <Loading />;
  // @ts-ignore
  if (error) return <Error />;

  return (
    <MenuItem ref={ref} component={Link} {...props} to={"/preferences"}>
      <ListItemIcon>
        <SettingsIcon />
      </ListItemIcon>
      <ListItemIcon>
        <ListItemText>Settings</ListItemText>
      </ListItemIcon>
    </MenuItem>
  );
});

const UserMenuAdmin = React.forwardRef((props: any, ref: any) => {
  const dataProvider = useDataProvider();

  const { data, isLoading, error } = useQuery(["preferences", "user"], () =>
    dataProvider.getOne("preferences", { id: "user" })
  );

  if (isLoading) return <Loading />;
  // @ts-ignore
  if (error) return <Error />;

  return (
    <MenuItem ref={ref} component={Link} {...props} to={"/admin"}>
      <ListItemIcon>
        <AdminPanelSettingsIcon />
      </ListItemIcon>
      <ListItemIcon>
        <ListItemText>Server administration</ListItemText>
      </ListItemIcon>
    </MenuItem>
  );
});

const UserMenu = (props: any) => {
  const authProvider = useAuthProvider();
  const isAdmin = authProvider.isAdmin();

  return (
    <RaUserMenu {...props}
    label="User"
    >
      <UserMenuPreferences />
      {isAdmin ? <UserMenuAdmin /> : null}
      <Logout />
    </RaUserMenu>
  );
};

export default UserMenu;
