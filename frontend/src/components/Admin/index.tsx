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

import { useEffect } from "react";

import {
  Datagrid,
  ListContextProvider,
  TextField,
  useListController,
  useShowController,
  BooleanField,
  useRecordContext,
  DateField,
  EditButton,
  Button,
  useAuthProvider,
} from "react-admin";
import {
  Box,
  Typography,
  Grid,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  useMediaQuery,
} from "@mui/material";

import AddIcon from "@mui/icons-material/Add";
import { ExpandMore } from "@mui/icons-material";
import {
  Processing,
  DocumentList,
  Runners,
  SearchEngineStatus,
} from "./Processing";
import { ByteToString } from "../util";
import { BooleanIndexingStatusField } from "../IndexingStatus";
import * as React from "react";
import { useNavigate } from "react-router-dom";
import { AuthPermissions } from "../../api/authProvider";

export const AdminView = (props: any) => {
  const { record, refetch, isLoading } = useShowController({
    ...props,
    resource: "admin",
    basePath: "/admin",
    id: "systeminfo",
  });

  const authProvider = useAuthProvider();
  const navigate = useNavigate();

  let interval: number = 0;

  useEffect(() => {
    // @ts-ignore
    interval = setInterval(() => {
      if (!isLoading) {
        refetch();
      }
    }, 5000);

    return function cleanup() {
      clearInterval(interval);
    };
  });

  authProvider.getPermissions({}).then((permissions: AuthPermissions) => {
    if (permissions.requiresReauthentication) {
      navigate("/auth/confirm-authentication");
    }
  });

  if (!record) return null;
  return (
    <>
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography variant="h5">Server info</Typography>
        </AccordionSummary>
        <AccordionDetails style={{ flexDirection: "column" }}>
          <Typography variant="h6">{record.name} </Typography>
          <Typography variant="h6">
            <a href="https://github.com/tryffel/virtualpaper">Home page</a>
          </Typography>
          <Typography color="textSecondary">
            Version: {record.version}, commit: {record.commit}{" "}
          </Typography>
          <Typography>Go version: {record.go_version} </Typography>
        </AccordionDetails>
      </Accordion>
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography variant="h5">Installation</Typography>
        </AccordionSummary>
        <AccordionDetails style={{ flexDirection: "column" }}>
          <Typography>Number of CPUs: {record.number_cpus} </Typography>
          <Typography>{record.imagemagick_version} </Typography>
          <Typography>
            Tesseract version: {record.tesseract_version}{" "}
          </Typography>
          <Typography>
            Poppler installed: {record.poppler_installed ? "Yes" : "No"}{" "}
          </Typography>
          <Typography>
            Pandoc installed: {record.pandoc_installed ? "Yes" : "No"}{" "}
          </Typography>
          <SearchEngineStatus status={record.search_engine_status} />
        </AccordionDetails>
      </Accordion>
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography variant="h5">Server statistics</Typography>
        </AccordionSummary>
        <AccordionDetails style={{ flexDirection: "column" }}>
          <Typography>Uptime: {record.uptime} </Typography>
          <Typography>Server load: {record.server_load} </Typography>
          <Typography>
            Total space used: {record.documents_total_size_string}{" "}
          </Typography>
          <Runners record={record} />
          <Typography variant="h5" marginTop={"1em"}>
            Documents statistics
          </Typography>
          <Typography>
            Waiting for processing: {record.documents_queued}{" "}
          </Typography>
          <Typography>
            Processed today: {record.documents_processed_today}{" "}
          </Typography>
          <Typography>
            Processed past week: {record.documents_processed_past_week}{" "}
          </Typography>
          <Typography>
            Processed past month: {record.documents_processed_past_month}{" "}
          </Typography>
          <Typography>Documents total: {record.documents_total} </Typography>
        </AccordionDetails>
      </Accordion>
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography variant="h5">Manage users</Typography>
        </AccordionSummary>
        <AccordionDetails style={{ flexDirection: "column" }}>
          <AddUserButton />
          <AdminShowUsers />
        </AccordionDetails>
      </Accordion>
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography variant="h5">Request Document processing</Typography>
        </AccordionSummary>
        <AccordionDetails style={{ flexDirection: "column" }}>
          <Processing />
        </AccordionDetails>
      </Accordion>
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMore />}>
          <Typography variant="h5">Document processing queue</Typography>
        </AccordionSummary>
        <AccordionDetails style={{ flexDirection: "column" }}>
          <DocumentList />
        </AccordionDetails>
      </Accordion>
    </>
  );
};

const AddUserButton = () => {
  const navigate = useNavigate();
  return (
    <Button onClick={() => navigate("/admin/users/create")} label={"Add user"}>
      <AddIcon />
    </Button>
  );
};

const AdminShowUsers = (props: any) => {
  const listContext = useListController({ ...props, resource: "admin/users" });
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  return (
    <ListContextProvider value={listContext}>
      <Datagrid expand={ShowExpandedUser}>
        <TextField source="id"></TextField>
        <TextField source="user_name"></TextField>
        {!isSmall ? <TextField source="email"></TextField> : null}
        <BooleanField source="is_active"></BooleanField>
      </Datagrid>
    </ListContextProvider>
  );
};

const ShowExpandedUser = () => {
  const record = useRecordContext();
  const documentsSize = record ? ByteToString(record.documents_size) : "0";

  return (
    <Grid container>
      <Grid item xs={6} md={6} lg={6}>
        <Box display={{ xs: "block", sm: "flex" }}>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <BooleanIndexingStatusField source="indexing" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Created at</Typography>
            <DateField source="created_at" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Updated at</Typography>
            <DateField source="updated_at" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Is administrator</Typography>
            <BooleanField source="is_admin" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2"># of documents</Typography>
            <TextField source="documents_count" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Storage size</Typography>
            <Typography variant="body1">{documentsSize}</Typography>
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <EditButton resource={"admin/users"} />
          </Box>
        </Box>
      </Grid>
    </Grid>
  );
};

export default AdminView;
