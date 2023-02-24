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
  Button,
  CreateButton,
  ExportButton,
  List,
  Pagination,
  SortButton,
  TopToolbar,
  useListContext,
  useStore,
} from "react-admin";
import {
  Grid,
  Toolbar as MuiToolbar,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";

import { HelpButton } from "../Help";
import { DocumentSearchFilter, FullTextSeachFilter } from "./SearchFilter";
import ClearIcon from "@mui/icons-material/Clear";
import EditIcon from "@mui/icons-material/Edit";
import { Link } from "react-router-dom";
import { DocumentCard } from "./DocumentCard";

const DocumentPagination = () => (
  <Pagination rowsPerPageOptions={[10, 25, 50, 100]} />
);

const DocumentListActions = () => (
  <TopToolbar>
    <DocumentHelp />
    <SortButton
      label="Sort"
      fields={["date", "name", "updated_at", "created_at"]}
    />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

export const DocumentList = () => {
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));
  if (isSmall) return <SmallDocumentList />;
  else return <LargeDocumentList />;
};

const SmallDocumentList = () => {
  return (
    <List
      title="All documents"
      actions={<DocumentListActions />}
      pagination={<DocumentPagination />}
    >
      <DocumentGrid />
    </List>
  );
};

const LargeDocumentList = () => {
  return (
    <List
      title="All documents"
      pagination={<DocumentPagination />}
      actions={<DocumentListActions />}
      sort={{ field: "date", order: "DESC" }}
      filters={[<DocumentSearchFilter />]}
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

  const clearSelected = () => {
    setSelectedIds([]);
  };

  const bulkEdit = () => {
    console.log("edit ids ", selectedIds);
  };

  return !isLoading && data ? (
    <Grid
      flex={2}
      sx={{
        background: theme.palette.background.default,
        padding: "1em",
      }}
    >
      <BulkEditToolbar
        selectedIds={selectedIds}
        clear={clearSelected}
        edit={bulkEdit}
      />
      <FullTextSeachFilter />
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
    <HelpButton title="Search documents">
      <Typography variant="h5" color="textPrimary">
        Full text search
      </Typography>
      <p>
        Query is single sentence that can consist of text, date range or
        metadata. Any single words or phrases are used to query the contents of
        the documents. Date range filters results by their date. Metadata
        filters results by metadata. The search bar uses autocomplete feature.
      </p>
      <p>
        For more help on search queries please see the official documentation.
      </p>

      <Typography variant="h5" color="textPrimary">
        Sample queries
      </Typography>

      <p>
        <Typography>Text</Typography>
        <Typography>- searching single words</Typography>
        <Typography>
          - searching single words AND (phrase must match)
        </Typography>
      </p>
      <p>
        <Typography>Metadata</Typography>
        <Typography>- key:value</Typography>
        <Typography>- author:doyle</Typography>
        <Typography>- (author:doyle OR author:christie)</Typography>
      </p>
      <p>
        <Typography>Date</Typography>
        <Typography>- date:today</Typography>
        <Typography>- date:2022</Typography>
        <Typography>- date:2015|2022 #(range between dates) </Typography>
        <Typography>- date:2015|today </Typography>
        <Typography>- date:2015-6-12|2022-8 </Typography>
      </p>

      <p>
        <Typography variant="h6" color="textPrimary">
          Combining multiple terms into single query
        </Typography>
        <Typography>
          Any term can be combined into more complex queries
        </Typography>
        <Typography>- word search date:2022 author:doyle </Typography>
        <Typography>
          - word search author:doyle (class:paper OR class:invoice)
        </Typography>
        <p>
          The last query is parsed as: "'word' AND 'search' AND
          metadata(author:doyle) AND (metadata(class:paper) OR
          metadata(class:invoice))"
        </p>
      </p>
    </HelpButton>
  );
};

const BulkEditToolbar = (props: any) => {
  const { selectedIds, clear, edit } = props;
  if (!selectedIds || !selectedIds.length) {
    return null;
  }

  const [store, setStore] = useStore("bulk-edit-document-ids", []);

  const onClear = () => {
    setStore([]);
    clear();
  };

  const onClick = () => {
    setStore(selectedIds);
  };

  return (
    <MuiToolbar
      sx={{
        backgroundColor: "rgb(230, 223, 243)",
        borderTopLeftRadius: "4px",
        borderTopRightRadius: "4px",
        paddingLeft: "24px",
        paddingRight: "24px",
        alignItems: "center",
        height: "48px",
        color: "#673ab7",
      }}
    >
      <Button onClick={onClear}>
        <ClearIcon />
      </Button>
      <Typography variant="body2" style={{ fontSize: "1em", flex: 1 }}>
        {selectedIds.length} Documents selected
      </Typography>
      <Button
        component={Link}
        // @ts-ignore
        to={{
          pathname: "/documents/bulkEdit/create/",
          search: `documents=${JSON.stringify(selectedIds)}`,
        }}
        label="Edit"
        style={{ fontSize: "1em" }}
        onClick={onClick}
      >
        <EditIcon />
      </Button>
    </MuiToolbar>
  );
};
