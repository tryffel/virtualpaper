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
  DateField,
  TextField,
  useAuthProvider,
  email,
  Labeled,
  SaveButton,
  Button,
  TabbedForm,
  Toolbar,
  SimpleForm,
} from "react-admin";

import {
  Typography,
  Grid,
  InputLabel,
  OutlinedInput,
  InputAdornment,
  IconButton,
  Tooltip,
} from "@mui/material";
import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";
import { Link } from "react-router-dom";
import { StopWordsInput, SynonymsInput } from "./Settings";
import KeyIcon from "@mui/icons-material/Key";

export const ProfileEdit = (...props: any) => {
  return (
    <Edit
      redirect={false}
      id="user"
      resource="preferences"
      basePath="/preferences"
      title="Profile"
      {...props}
    >
      <SimpleForm
        warnWhenUnsavedChanges
        toolbar={
          <Toolbar>
            <SaveButton /> <ResetPasswordButton />
          </Toolbar>
        }
      >
        <Grid container spacing={2}>
          <Grid item xs={12} lg={6}>
            <UserBasicInfo />
          </Grid>
          <Grid item xs={12} lg={6}>
            <Statistics />
          </Grid>
        </Grid>
      </SimpleForm>
    </Edit>
  );
};

const ResetPasswordButton = () => {
  return (
    <Link
      to={"/auth/forgot-password"}
      style={{
        fontSize: 16,
        textDecoration: "none",
        paddingLeft: "5px",
        marginLeft: "5px",
      }}
    >
      <Button size="small" label={"Reset password"}>
        <KeyIcon />
      </Button>
    </Link>
  );
};

const ShowToken = () => {
  const authProvider = useAuthProvider();
  const token = authProvider.getToken();
  const [tokenShown, setTokenShown] = React.useState(false);

  const handleClickShowPassword = () => {
    setTokenShown(!tokenShown);
  };

  const handleMouseDownPassword = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    event.preventDefault();
  };

  return (
    <>
      <InputLabel htmlFor="outlined-adornment-password">API Token</InputLabel>
      <Tooltip title="Api token. Please read documentation first. This will grant access to all user data, so please be careful not to expose it.">
        <OutlinedInput
          multiline
          id="outlined-adornment-password"
          type={tokenShown ? "text" : "password"}
          value={tokenShown ? token : "******"}
          color={tokenShown ? "warning" : "primary"}
          readOnly
          fullWidth
          endAdornment={
            <InputAdornment position="end">
              <IconButton
                aria-label="toggle password visibility"
                onClick={handleClickShowPassword}
                onMouseDown={handleMouseDownPassword}
                edge="end"
              >
                {tokenShown ? <VisibilityOff /> : <Visibility />}
              </IconButton>
            </InputAdornment>
          }
          label="Password"
        />
      </Tooltip>
    </>
  );
};

const UserBasicInfo = () => {
  return (
    <Grid container spacing={1}>
      <Grid item xs={12}>
        <Typography variant="h6">User settings</Typography>
      </Grid>

      <Grid item xs={6}>
        <Labeled label="User id">
          <TextField source="user_id" />
        </Labeled>
      </Grid>
      <Grid item xs={6}>
        <Labeled label="Username">
          <TextField source="user_name" />
        </Labeled>
      </Grid>
      <Grid item xs={12} md={8}>
        <TextInput source="email" validate={email()} />
      </Grid>

      <Grid item xs={12} md={8}>
        <ShowToken />
      </Grid>
    </Grid>
  );
};

const Statistics = () => {
  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Typography variant="h5">Statistics</Typography>
      </Grid>
      <Grid item xs={12} sm={6}>
        <Labeled label="Number of documents">
          <TextField source="documents_count" label={"Documents count"} />
        </Labeled>
      </Grid>
      <Grid item xs={12} sm={6}>
        <Labeled label="Total size of all documents">
          <TextField
            source="documents_size_string"
            label={"Total size of documents"}
          />
        </Labeled>
      </Grid>
      <Grid item xs={6}>
        <Labeled label="Joined at">
          <DateField source="created_at" />
        </Labeled>
      </Grid>
      <Grid item xs={6}>
        <Labeled label="Updated at">
          <DateField source="updated_at" />
        </Labeled>
      </Grid>
    </Grid>
  );
};

const StopWordsTab = () => {
  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant={"h5"}>Search experience: stop words</Typography>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="body2">
          Stop words are language-specific words that are excluded when
          documents during search query. They will not modify the documents in
          any way, they are only meant to improve the relevancy of search
          results.
        </Typography>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="body2">Format: one stop word per line</Typography>
        <StopWordsInput />
      </Grid>
    </Grid>
  );
};

const SynonymsTab = () => {
  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant={"h5"}>Search experience: synonyms</Typography>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="body2">
          Synonyms are words that are treated as same when searching documents.
          They will not modify the contents of documents in any way, they will
          only improve the relevancy of the search results.
        </Typography>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="body2">
          Format: list of synonyms separated by comma, e.g.
          ('food','spaghetti','pasta')
        </Typography>
        <SynonymsInput />
      </Grid>
    </Grid>
  );
};

export default {
  edit: ProfileEdit,
};
