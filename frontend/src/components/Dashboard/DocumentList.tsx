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

import * as React from "react";

import { Loading, useGetMany } from "react-admin";

import { useTheme, Grid, Typography, Box, Paper } from "@mui/material";
import { isDOMComponent } from "react-dom/test-utils";
import { DocumentCard } from "../Documents/List";

export const LastUpdatedDocumentList = (props: { ids: string[] }) => {
  const theme = useTheme();
  const { ids } = props;
  const { data, isLoading, error, refetch } = useGetMany("documents", {
    ids: ids.slice(0, 5),
  });

  if (isLoading) {
    return <Loading />;
  }

  if (data) {
    return (
      <Paper elevation={2}>
        <Typography variant="h5" gutterBottom marginLeft="1em">
          Latest documents
        </Typography>
        {data.map((document) => (
          <DocumentCard record={document} />
        ))}
      </Paper>
    );
  }
  return null;
};
