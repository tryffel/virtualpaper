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
  List,
  Datagrid,
  ChipField,
  TextField,
  DateField,
  NumberField,
} from "react-admin";

import {
  Container,
  Grid,
  Paper,
  Typography,
  useMediaQuery,
} from "@mui/material";
import { EmptyResourcePage } from "../../components/primitives/EmptyPage";
import React from "react";
import {
  SearchMetadataField,
  SearchMetadataResult,
} from "@resources/MetadataKeys/Search.tsx";
import debounce from "lodash/debounce";

export const MetadataKeyList = () => {
  const [rawQuery, setRawQuery] = React.useState("");
  const [debouncedQuery, setDebouncedQuery] = React.useState("");

  const debounced = React.useMemo(
    () =>
      debounce((value) => {
        console.log("debouncing", value);
        setDebouncedQuery(value);
      }, 100),
    [],
  );
  const setQuery = (value: string) => {
    setRawQuery(value);
    debounced(value);
  };

  return (
    <Container>
      <Paper>
        <Grid container>
          <Grid item sx={{ ml: "auto", mr: "10px", mt: 1 }}>
            <SearchMetadataField query={rawQuery} setQuery={setQuery} />
          </Grid>
          <Grid item xs={12}>
            <ListResults query={debouncedQuery} setQuery={setQuery} />
          </Grid>
        </Grid>
      </Paper>
    </Container>
  );
};

const ListResults = ({
  query,
  setQuery,
}: {
  query: string;
  setQuery: (query: string) => void;
}) => {
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));
  if (query === "") {
    return (
      <Grid container spacing={1}>
        <Grid item xs={12} sx={{ ml: 1, mr: 1, mb: 2 }}>
          <Typography variant={"h5"}>Keys</Typography>
        </Grid>
        <Grid item xs={12}>
          <List
            title="Metadata"
            sort={{ field: "key", order: "ASC" }}
            empty={<EmptyMetadataList />}
            actions={false}
          >
            <Datagrid rowClick="edit" bulkActionButtons={false}>
              <ChipField source="key" label={"Name"} />
              <TextField source="comment" label={"Description"} />
              {!isSmall ? (
                <DateField source="created_at" label={"Created at"} />
              ) : null}
              <NumberField
                source="metadata_values_count"
                label={"Total keys"}
              />
              <NumberField source="documents_count" label={"Total documents"} />
            </Datagrid>
          </List>
        </Grid>
      </Grid>
    );
  }

  return <SearchMetadataResult query={query} setQuery={setQuery} />;
};

const EmptyMetadataList = () => {
  return (
    <EmptyResourcePage
      title={"No metadata keys"}
      subTitle={"Do you want to add one?"}
    />
  );
};
