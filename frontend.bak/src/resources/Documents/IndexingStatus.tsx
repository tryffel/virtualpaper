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
import { Labeled, useRecordContext } from "react-admin";
import { Box, CircularProgress, Typography } from "@mui/material";

export const IndexingStatusField = (props: any) => {
  const record = useRecordContext(props);

  const [value, setValue] = React.useState("ready");
  const [label, setLabel] = React.useState("Ready");

  if (props.record && props.source) {
    const status = props.record[props.source];
    if (status === "indexing" && value !== status) {
      setValue(status);
      setLabel("Processing ongoing");
    } else if (status === "pending" && value !== status) {
      setValue(status);
      setLabel("Waiting for processing to start");
    } else if (status === "ready" && value !== status) {
      setValue(status);
      setLabel("Document processed successfully");
    }
  }

  return value === "ready" ? null : (
    <Box flex={0} mr={{ xs: 0, sm: "0.5em" }}>
      <Labeled label="Document processing status">
        <>
          <CircularProgress
            variant="indeterminate"
            size={25}
            color="secondary"
            {...props}
          />
          <Typography variant="caption" component="div" color="textSecondary">
            {label}
          </Typography>
        </>
      </Labeled>
    </Box>
  );
};
