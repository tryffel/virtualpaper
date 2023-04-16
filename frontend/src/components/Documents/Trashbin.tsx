/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
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
import { useState } from "react";
import {
  List,
  Loading,
  Pagination,
  SortButton,
  TopToolbar,
  useListContext,
} from "react-admin";
import { Grid, Typography, useMediaQuery, useTheme } from "@mui/material";
import DeleteOutlinedIcon from "@mui/icons-material/DeleteOutlined";

import { HelpButton } from "../Help";
import { DocumentSearchFilter, FullTextSeachFilter } from "./SearchFilter";
import { DocumentCard } from "./DocumentCard";
import { EmptyResourcePage } from "../primitives/EmptyPage";

const DocumentPagination = () => (
  <Pagination rowsPerPageOptions={[10, 25, 50, 100]} />
);

const DocumentListActions = () => (
  <TopToolbar>
    <DocumentHelp />
    <SortButton
      label="Sort"
      fields={["date", "name", "updated_at", "created_at", "deleted_at"]}
    />
  </TopToolbar>
);

export const DeletedDocumentList = () => {
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));
  if (isSmall) return <SmallDocumentList />;
  else return <LargeDocumentList />;
};

const SmallDocumentList = () => {
  return (
    <List
      title="Documents"
      actions={<DocumentListActions />}
      sort={{ field: "deleted_at", order: "DESC" }}
      pagination={<DocumentPagination />}
      empty={<EmptyTrashBin />}
    >
      <DocumentGrid />
    </List>
  );
};

const LargeDocumentList = () => {
  return (
    <List
      title="Documents"
      pagination={<DocumentPagination />}
      actions={<DocumentListActions />}
      sort={{ field: "deleted_at", order: "DESC" }}
      filters={[<DocumentSearchFilter />]}
      empty={<EmptyTrashBin />}
    >
      <DocumentGrid />
    </List>
  );
};

const DocumentGrid = (props: any) => {
  const { data, isLoading } = useListContext();
  const theme = useTheme();

  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  const isSelected = (id: string) => {
    const found = selectedIds.includes(id);
    return found;
  };

  const toggleSelectedId = (id: string) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter((item) => item != id));
    } else {
      setSelectedIds(selectedIds.concat([id]));
    }
  };

  if (isLoading) {
    return <Loading />;
  }

  return data ? (
    <Grid
      flex={2}
      sx={{
        background: theme.palette.background.default,
        padding: "1em",
      }}
    >
      <Typography variant="h6">Document trash bin</Typography>
      {data.map((record) => (
        <DocumentCard
          record={record}
          selected={isSelected}
          setSelected={toggleSelectedId}
        />
      ))}
    </Grid>
  ) : null;
};

const DocumentHelp = () => {
  return (
    <HelpButton title="Deleted documents">
      <p>
        When document is deleted from the server, the document is moved here,
        allowing user to restore the document.
      </p>
      <p>
        Documents will be permanently deleted after 14 days by default (or as
        configured in the configuration file).
      </p>

      <p>
        Document can also be permanently deleted by clicking the Delete button.
      </p>
    </HelpButton>
  );
};

const EmptyTrashBin = () => {
  return (
    <EmptyResourcePage
      title={"Trash bin is empty"}
      noCreateButton={true}
      subTitle={""}
      icon={<DeleteOutlinedIcon />}
    />
  );
};
