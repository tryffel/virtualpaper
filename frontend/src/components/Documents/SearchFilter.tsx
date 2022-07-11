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

import { Card, CardContent, Box } from "@mui/material";
import { FilterList, FilterListItem, TextInput } from "react-admin";
import {
  endOfYesterday,
  startOfMonth,
  startOfWeek,
  startOfYear,
  subYears,
} from "date-fns";
import { AccessTime } from "@mui/icons-material";
import * as React from "react";

export const DocumentSearchFilter = [
  <TextInput
    source="q"
    label="Full Text Search"
    alwaysOn
    resettable
    fullWidth
  />,
  <TextInput
    label="Metadata filter"
    alwaysOn
    resettable
    fullWidth
    source="metadata"
  />,
];

export const FilterSidebar = () => (
  <Box
    sx={{
      display: {
        xs: "none",
        sm: "block",
      },
      order: -1, // display on the left rather than on the right of the list
      width: "15em",
      marginRight: "1em",
    }}
  >
    <Card style={{ width: "13em", minWidth: "13em" }}>
      <CardContent>
        <LastVisitedFilter />
      </CardContent>
    </Card>
  </Box>
);

const LastVisitedFilter = () => (
  <FilterList label="Document date" icon={<AccessTime />}>
    <FilterListItem
      label="Today"
      value={{
        after: endOfYesterday().getTime(),
        before: undefined,
      }}
    />
    <FilterListItem
      label="This week"
      value={{
        after: startOfWeek(new Date()).getTime(),
        before: undefined,
      }}
    />
    <FilterListItem
      label="This month"
      value={{
        after: startOfMonth(new Date()).getTime(),
        before: undefined,
      }}
    />
    <FilterListItem
      label="This year"
      value={{
        after: startOfYear(new Date()).getTime(),
        before: undefined,
      }}
    />
    <FilterListItem
      label="Last year"
      value={{
        after: subYears(startOfYear(new Date()), 1).getTime(),
        before: startOfYear(new Date()).getTime(),
      }}
    />
    <FilterListItem
      label="Earlier"
      value={{
        after: undefined,
        before: subYears(startOfYear(new Date()), 2).getTime(),
      }}
    />
  </FilterList>
);
