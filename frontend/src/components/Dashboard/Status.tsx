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

import { Grid, Typography, Box, CircularProgress } from "@mui/material";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";

export const IndexingStatus = (props: { indexing: boolean }) => {
  const { indexing } = props;

  if (indexing) {
    return (
      <Grid container>
        <Grid margin="1em">
          <CircularProgress
            variant="indeterminate"
            size={30}
            color="secondary"
            {...props}
          />
        </Grid>
        <Grid>
          <Typography variant="body1" color="textSecondary" marginTop="1em">
            Documents indexing in progress
          </Typography>
        </Grid>
      </Grid>
    );
  } else {
    return (
      <Grid container>
        <Grid margin="1em">
          <CheckCircleIcon color="primary" />
        </Grid>
        <Grid>
          <Typography variant="body1" color="textSecondary" marginTop="1em">
            All documents indexed
          </Typography>
        </Grid>
      </Grid>
    );
  }
};
