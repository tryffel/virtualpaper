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
  RecordContextProvider,
  useAuthProvider,
  useShowController,
} from "react-admin";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Container,
  Grid,
  Typography,
} from "@mui/material";

import ExpandMore from "@mui/icons-material/ExpandMore";
import {
  DocumentList,
  Processing,
  Runners,
  SearchEngineStatus,
} from "./Processing";
import { useNavigate } from "react-router-dom";
import { AuthPermissions } from "../../api/authProvider";
import {
  BuildInfo,
  InstallationInfo,
  ProcessingStatistics,
  ServerStatistics,
} from "./ServerInfo";
import { AdminShowUsers } from "./UserList";

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
    <Container>
      <RecordContextProvider value={record}>
        <Accordion>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography variant="h5">System information</Typography>
          </AccordionSummary>
          <AccordionDetails style={{ flexDirection: "column" }}>
            <Grid container spacing={4}>
              <Grid item xs={12} md={6}>
                <BuildInfo />
              </Grid>
              <Grid item xs={12} md={6}>
                <InstallationInfo />
              </Grid>
              <Grid item xs={12} md={6}>
                <ServerStatistics />
              </Grid>
              <Grid item xs={12} md={6}>
                <ProcessingStatistics />
              </Grid>
              <Grid item xs={12} marginTop={1}>
                <Runners />
              </Grid>
              <Grid item xs={12} marginTop={1}>
                <SearchEngineStatus />
              </Grid>
            </Grid>
          </AccordionDetails>
        </Accordion>
        <Accordion>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography variant="h5">Manage users</Typography>
          </AccordionSummary>
          <AccordionDetails style={{ flexDirection: "column" }}>
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
      </RecordContextProvider>
    </Container>
  );
};

export default AdminView;
