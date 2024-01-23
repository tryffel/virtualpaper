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
  Autocomplete,
  Grid,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { DesktopDatePicker } from "@mui/x-date-pickers/DesktopDatePicker";
import { LocalizationProvider } from "@mui/x-date-pickers/LocalizationProvider";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import {
  FilterList,
  FilterListItem,
  TextInput,
  useDataProvider,
  useListContext,
  useListFilterContext,
  useTheme,
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
import { debounce, filter, throttle, values } from "lodash";
import { string } from "prop-types";

export const DocumentSearchFilter = () => {
  return (
    <TextInput
      source="q"
      label="Search"
      alwaysOn
      resettable
      fullWidth
      sx={{
        minWidth: "40em",
      }}
    />
  );
};

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
      let values = { ...filterValues };
      if (props.mode === "after") {
        values = values && { after: newValue ? newValue.getTime() : null };
      } else {
        values = values && { before: newValue ? newValue.getTime() : null };
      }
      setFilters({ ...values }, null, false);
    }
  };

  const onReset = () => {
    let values = {};
    if (props.mode === "after") {
      values = values && { after: value ? value.getTime() : null };
    } else {
      values = values && { before: value ? value.getTime() : null };
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
      <Typography variant="body1">
        {props.mode === "after" ? "After" : "Before"}
      </Typography>
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

interface Suggestion {
  value: string;
  type: string;
  hint: string;
  prefixed: string;
}

export const FullTextSeachFilter = (props: any) => {
  const theme = useTheme();

  const { filterValues, setFilters } = useListContext();
  const addFilter = (q: string) => {
    setFilters({ ...filterValues, q: q }, false);
  };

  const removeFilter = () => {
    //const keysToRemove = Object.keys(value);
    const keyToRemove = "q";
    const filters = Object.keys(filterValues).reduce(
      (acc, key) => (key === "q" ? acc : { ...acc, [key]: filterValues[key] }),
      {}
    );

    setFilters(filters, null, false);
  };

  //const navigate = useNavigate();
  const [options, setOptions] = React.useState<Suggestion[]>([]);
  const [query, setQuery] = React.useState(filterValues.q);

  const [prefix, setPrefix] = React.useState(filterValues.q);

  const [value, setValue] = React.useState<Suggestion>({
    value: "",
    type: "",
    hint: "",
    prefixed: filterValues.q ? filterValues.q : "",
  });

  const [validQuery, setValidQuery] = React.useState(true);
  const [initial, setInitial] = React.useState(filterValues.q === undefined);

  const dataProvider = useDataProvider();

  const getSuggestions = (q: string) => {
    dataProvider
      .suggestSearch({
        data: {
          filter: q,
        },
      })
      // @ts-ignore
      .then((data) => {
        // @ts-ignore
        setOptions(
          Object.values(data.data.suggestions).map(
            // @ts-ignore
            (val: Suggestion) => ({
              type: val.type,
              hint: val.hint,
              prefixed: data.data.prefix.trimStart() + val.value,
              value: val.value,
            })
          )
        );
        setPrefix(data.data.prefix);
        setValidQuery(data.data.validQuery);
        if (!initial) {
          setInitial(true);
        }
      });
  };

  const doNavigate = (q: string) => {
    if (q === "") {
    removeFilter();
    } else {
    addFilter(q);
    }
  };

  const throttledSuggestions = React.useCallback(
    debounce(getSuggestions, 100, { leading: true, trailing: false }),
    [query]
  );

  const throttledNavigate = React.useMemo(
    () => debounce(doNavigate, 500, { leading: false, trailing: true }),
    [query]
  );

  React.useEffect(() => {
    throttledSuggestions(query);
    throttledNavigate(query);

    return () => {
      throttledSuggestions.cancel();
      throttledNavigate.cancel();
    };
  }, [query, initial]);

  const styleForType = (type: string) => {
    if (type === "key") {
      return { fontStyle: "italic" };
    }
    if (type === "metadata") {
      return {};
    }
    if (type === "primary") {
      return {};
    }
  };

  const colors = {
    key: "",
    metadata: "primary",
    operand: "",
  };

  return (
    <Autocomplete
      disablePortal
      id="full-text-search"
      filterOptions={(x) => x}
      options={options}
      autoComplete
      includeInputInList
      value={value}
      // @ts-ignore
      onChange={(event: any, newValue: Suggestion) => {
        // use selected suggestion
        setQuery(newValue.prefixed);
        setValue(newValue);
      }}
      onInputChange={(event, input) => {
        // user entered text
        setQuery(input);
        setValue({ value: input, type: "", hint: "", prefixed: input });
      }}
      isOptionEqualToValue={(option, value): boolean => {
        // set to never equal to ensure list of options is always visible
        return false;
      }}
      renderInput={(params) => (
        <TextField {...params} label="Search documents" fullWidth />
      )}
      getOptionLabel={(option) => option.prefixed}
      renderOption={(props, option) => {
        return (
          <li {...props}>
            <Grid
              container
              spacing={0}
              alignItems="left"
              justifyContent="flex-start"
            >
              <Grid item xs="auto" mr={0}>
                <Typography variant="body1">
                  {prefix.slice(-1) === " "
                    ? prefix.trimEnd() + "\u00A0"
                    : prefix}
                </Typography>
              </Grid>
              <Grid item xs="auto" mr={0}>
                <Typography
                  variant="body1"
                  // @ts-ignore
                  color={colors[option.type]}
                  style={styleForType(option.type)}
                >
                  {option.value}
                </Typography>
              </Grid>
            </Grid>
          </li>
        );
      }}
    />
  );
};
