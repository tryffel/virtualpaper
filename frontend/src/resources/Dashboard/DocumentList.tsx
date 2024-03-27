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
import { Loading, useGetMany, useStore } from "react-admin";
import {
  Typography,
  Box,
  Paper,
  ToggleButtonGroup,
  ToggleButton,
  useMediaQuery,
} from "@mui/material";
import { DocumentCard } from "@components/document/card";

export const LastUpdatedDocumentList = (props: {
  lastUpdatedIds: string[];
  lastAddedIds: string[];
  lastViewedIds: string[];
  favorites: string[];
}) => {
  const isNotSmall = useMediaQuery((theme: any) => theme.breakpoints.up("xs"));
  const { lastUpdatedIds, lastAddedIds, lastViewedIds, favorites } = props;
  const [showMode, setShowMode] = useStore<ShowMode>(
    "dashboard.latest_documents.mode",
    "lastUpdated",
  );

  const getDocumentIds = () => {
    switch (showMode) {
      case "lastUpdated":
        return lastUpdatedIds;
      case "lastAdded":
        return lastAddedIds;
      case "lastViewed":
        return lastViewedIds;
      case "favorites":
        return favorites;
      default:
        return lastUpdatedIds;
    }
  };

  const { data, isLoading } = useGetMany("documents", {
    ids: getDocumentIds()?.slice(0, 10) ?? [],
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
            flexDirection: "column",
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

type ShowMode = "lastUpdated" | "lastAdded" | "lastViewed" | "favorites";

const ShowModeButton = (props: {
  showMode: ShowMode;
  setShowMode: (mode: ShowMode) => void;
}) => {
  const { showMode, setShowMode } = props;

  const handleAlignment = (
    _event: React.MouseEvent<HTMLElement>,
    newAlignment: ShowMode,
  ) => {
    newAlignment && setShowMode(newAlignment);
  };

  return (
    <ToggleButtonGroup
      value={showMode}
      exclusive
      onChange={handleAlignment}
      sx={{ pr: 1 }}
    >
      <ToggleButton size="small" value="favorites" color="primary">
        Favorites
      </ToggleButton>
      <ToggleButton size="small" value="lastAdded" color="primary">
        Added
      </ToggleButton>
      <ToggleButton size="small" value="lastUpdated" color="primary">
        Updated
      </ToggleButton>
      <ToggleButton size="small" value="lastViewed" color="primary">
        Viewed
      </ToggleButton>
    </ToggleButtonGroup>
  );
};
