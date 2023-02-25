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

import {
  useTheme,
  Typography,
  Box,
  Paper,
  ToggleButtonGroup,
  ToggleButton,
  useMediaQuery,
} from "@mui/material";
import { DocumentCard } from "../Documents/DocumentCard";

export const LastUpdatedDocumentList = (props: {
  lastUpdatedIds: string[];
  lastAddedIds: string[];
}) => {
  const theme = useTheme();
  const isNotSmall = useMediaQuery((theme: any) => theme.breakpoints.up("sm"));
  const { lastUpdatedIds, lastAddedIds } = props;
  const [showMode, setShowMode] = React.useState<ShowMode>("lastUpdated");
  const { data, isLoading, error, refetch } = useGetMany("documents", {
    ids:
      showMode === "lastUpdated"
        ? lastUpdatedIds
          ? lastUpdatedIds.slice(0, 5)
          : []
        : lastAddedIds
        ? lastAddedIds.slice(0, 5)
        : [],
  });

  if (isLoading) {
    return <Loading />;
  }

  if (props.lastUpdatedIds && data) {
    return (
      <Paper elevation={2}>
        <Box
          sx={{
            pt: 2,
            pb: 2,
            display: "flex",
            flexDirection: "row",
            justifyContent: "space-between",
          }}
        >
          <Typography variant="h5" gutterBottom sx={{ ml: 2, mr: 1 }}>
            Latest documents
          </Typography>
          {isNotSmall && (
            <ShowModeButton showMode={showMode} setShowMode={setShowMode} />
          )}
        </Box>
        <Box sx={{ pt: 2, pb: 2 }}>
          {data.map((document) => (
            <DocumentCard record={document} />
          ))}
        </Box>
      </Paper>
    );
  }
  return null;
};

type ShowMode = "lastUpdated" | "lastAdded";

const ShowModeButton = (props: {
  showMode: ShowMode;
  setShowMode: (mode: ShowMode) => void;
}) => {
  const { showMode, setShowMode } = props;

  const handleAlignment = (
    event: React.MouseEvent<HTMLElement>,
    newAlignment: ShowMode
  ) => {
    setShowMode(newAlignment);
  };

  return (
    <ToggleButtonGroup
      value={showMode}
      exclusive
      onChange={handleAlignment}
      sx={{ pr: 1 }}
    >
      <ToggleButton size="small" value="lastAdded">
        Updated
      </ToggleButton>
      <ToggleButton size="small" value="lastUpdated">
        Added
      </ToggleButton>
    </ToggleButtonGroup>
  );
};
