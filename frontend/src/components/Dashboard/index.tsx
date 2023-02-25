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

import { useGetOne, Loading, CreateButton } from "react-admin";
import { Box, Grid, Typography, Container } from "@mui/material";

import { Stats } from "./Stats";
import { LastUpdatedDocumentList } from "./DocumentList";
import { get } from "lodash";
import { IndexingStatusRow, IndexingStatusRowSmall } from "./Status";
import * as React from "react";

export const Dashboard = () => {
  const { data, isLoading } = useGetOne("documents/stats", {
    id: "",
  });
  if (isLoading) {
    return <Loading />;
  }

  const direction = "column";

  if (data.num_documents > 0) {
    return (
      <Grid spacing={1} direction={direction} flexGrow={1} alignItems="stretch">
        {/* <Grid item xl={2} lg={5} xs={12} sm={10}>
                    <DocumentTimeline stats={data.yearly_stats}/>
                </Grid> */}
        <Grid item xs={4} sm={4} md={4} lg={12}>
          <IndexingStatusRow indexing={get(data, "indexing")} />
        </Grid>

        <Grid item xs={12} sm={10} md={8} lg={3}>
          <LastUpdatedDocumentList
            lastUpdatedIds={get(data, "last_documents_updated")}
            lastAddedIds={get(data, "last_documents_added")}
          />
        </Grid>
      </Grid>
    );
  } else {
    return <EmptyDashboard />;
  }
};

export const ShowDocumentsIndexing = () => {
  const { data, isLoading } = useGetOne("documents/stats", {
    id: "",
  });
  if (isLoading) {
    return null;
  }
  return (
    <IndexingStatusRowSmall indexing={get(data, "indexing")} hideReady={true} />
  );
};

const EmptyDashboard = () => {
  return (
    <Grid
      container
      direction="column"
      justifyContent="center"
      alignItems="center"
    >
      <Grid item md={5}>
        <Box sx={{ mt: 10 }}>
          <Typography variant="h4" paragraph>
            No documents yet
          </Typography>
          <Typography variant="body1" paragraph>
            Upload some documents first!
          </Typography>
          <CreateButton resource="documents" />
        </Box>
      </Grid>
    </Grid>
  );
};
