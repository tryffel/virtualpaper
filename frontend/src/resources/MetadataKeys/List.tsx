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

import {
  List,
  Datagrid,
  ChipField,
  TextField,
  DateField,
  NumberField,
} from "react-admin";

import { useMediaQuery } from "@mui/material";
import { EmptyResourcePage } from "../../components/primitives/EmptyPage";

export const MetadataKeyList = () => {
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  return (
    <List
      title="Metadata"
      sort={{ field: "key", order: "ASC" }}
      empty={<EmptyMetadataList />}
    >
      <Datagrid rowClick="edit" bulkActionButtons={false}>
        <ChipField source="key" label={"Name"} />
        <TextField source="comment" label={"Description"} />
        {!isSmall ? (
          <DateField source="created_at" label={"Created at"} />
        ) : null}
        <NumberField source="metadata_values_count" label={"Total keys"} />
        <NumberField source="documents_count" label={"Total documents"} />
      </Datagrid>
    </List>
  );
};

const EmptyMetadataList = () => {
  return (
    <EmptyResourcePage
      title={"No metadata keys"}
      subTitle={"Do you want to add one?"}
    />
  );
};
