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
  Edit,
  TextInput,
  SimpleForm,
  DateField,
  TextField,
  DateInput,
} from "react-admin";

import { Box, Button, Typography, Grid } from "@mui/material";
import { JsonInput } from "react-admin-json-view";

export const ProfileEdit = (staticContext: any, ...props: any) => {
  return (
    <Edit
      redirect={false}
      id="user"
      resource="preferences"
      basePath="/preferences"
      title="Profile"
      {...props}
    >
      <SimpleForm>
        <div>
          <Grid container width={{ xs: "100%", xl: 800 }} spacing={2}>
            <Grid item xs={12} md={8}>
              <Typography variant="h5">Basic info</Typography>
              <TextInput disabled source="user_id" label={"User Id"} />

              <Box display={{ xs: "block", sm: "flex" }}>
                <Box mr={{ xs: 0, sm: "0.5em" }}>
                  <TextInput source="email" />
                </Box>
                <Box mr={{ xs: 0, sm: "0.5em" }}>
                  <TextInput disabled source="user_name" label={"Username"} />
                </Box>
              </Box>
              <Box display={{ xs: "block", sm: "flex" }}>
                <Box mr={{ xs: 0, sm: "0.5em" }}>
                  <DateInput disabled source="created_at" label={"Joined at"} />
                </Box>
                <Box mr={{ xs: 0, sm: "0.5em" }}>
                  <DateInput
                    disabled
                    source="updated_at"
                    label={"Last updated"}
                  />
                </Box>
              </Box>
            </Grid>
            <Grid item xs={12} md={8}>
              <Typography variant="h5">Statistics</Typography>
              <Box flex={1} ml={{ xs: 0, sm: "0.5em" }}>
                <Box mr={{ xs: 0, sm: "0.5em" }}>
                  <Typography variant="h6">Total documents</Typography>

                  <TextField
                    source="documents_count"
                    label={"Documents count"}
                  />
                </Box>
                <Box mr={{ xs: 0, sm: "0.5em" }}>
                  <Typography variant="h6">Total size</Typography>
                  <TextField
                    source="documents_size_string"
                    label={"Total size of documents"}
                  />
                </Box>
              </Box>
            </Grid>
          </Grid>
        </div>
      </SimpleForm>
    </Edit>
  );
};

export default {
  edit: ProfileEdit,
};
