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

import { LastUpdatedDocumentList } from "./DocumentList";
import { get } from "lodash";
import { IndexingStatusRow, IndexingStatusRowSmall } from "./Status";
import * as React from "react";
import { EmptyDocumentList } from "../Documents/List";

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
        <Grid
          container
          xs={8}
          sm={6}
          md={12}
          lg={12}
          direction={"row"}
          justifyContent={"space-between"}
        >
          <Grid item>
            <IndexingStatusRow indexing={get(data, "indexing")} />
          </Grid>
          <Grid item sx={{ mt: "14px" }}>
            <CreateButton resource={"documents"} label={"Upload document"} />
          </Grid>
        </Grid>

        <Grid item xs={12} sm={10} md={8} lg={3}>
          <LastUpdatedDocumentList
            lastUpdatedIds={get(data, "last_documents_updated")}
            lastAddedIds={get(data, "last_documents_added")}
            lastViewedIds={get(data, "last_documents_viewed")}
          />
        </Grid>
      </Grid>
    );
  } else {
    return <EmptyDocumentList />;
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
