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

import {
  Card,
  CardContent,
  Box,
  TextField,
  Typography,
  Button,
} from "@mui/material";
import { DesktopDatePicker } from "@mui/x-date-pickers/DesktopDatePicker";
import { LocalizationProvider } from "@mui/x-date-pickers/LocalizationProvider";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import {
  FilterList,
  FilterListItem,
  TextInput,
  useListFilterContext,
} from "react-admin";
import {
  endOfYesterday,
  startOfMonth,
  startOfWeek,
  startOfYear,
  subYears,
} from "date-fns";
import { AccessTime } from "@mui/icons-material";
import * as React from "react";
import { values } from "lodash";

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
        <DatePicker mode="after" />
        <DatePicker mode="before" />
      </CardContent>
    </Card>
  </Box>
);

// @ts-ignore
const DatePicker = (props) => {
  const { filterValues, setFilters } = useListFilterContext();
  const [value, setValue] = React.useState<Date | null>(null);

  const handleChange = (newValue: Date | null) => {
    setValue(newValue);
    if (newValue) {
    let values = {...filterValues};
    if (props.mode === "after") {
      values = values && {after: newValue? newValue.getTime(): null}
    } else {
      values = values && {before: newValue? newValue.getTime(): null}
    }
    setFilters({ ...values}, null, false);
  }
  };

  const onReset = () => {
    let values = {}
    if (props.mode === "after") {
      values = values && {after: value? value.getTime(): null}
    } else {
      values = values && {before: value? value.getTime(): null}
    }
    const keysToRemove = Object.keys(values);
    const filters = Object.keys(filterValues).reduce(
      (acc, key) =>
        keysToRemove.includes(key) ? acc : { ...acc, [key]: filterValues[key] },
      {}
    );
    setFilters(filters, null, false);
    setValue(null);
  };

  return (
    <Box>
      <Typography variant="body1">{props.mode === "after" ? "After": "Before"}</Typography>
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DesktopDatePicker
          onChange={handleChange}
          inputFormat="MM/dd/yyyy"
          value={value}
          renderInput={(params) => <TextField {...params} />}
        />
      </LocalizationProvider>
      <Button onClick={onReset}>Reset</Button>
    </Box>
  );
};

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
