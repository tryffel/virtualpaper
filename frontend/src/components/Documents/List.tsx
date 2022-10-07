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
import {
  DateField,
  EditButton,
  List,
  ShowButton,
  useListContext,
  Pagination,
  TopToolbar,
  SortButton,
  ExportButton,
  CreateButton,
  Button,
  useStore,
} from "react-admin";
import {
  useMediaQuery,
  Grid,
  Card,
  CardContent,
  CardActions,
  CardHeader,
  Typography,
  Box,
  useTheme,
  ToggleButton,
  Toolbar as MuiToolbar,
  IconButton,
} from "@mui/material";

import get from "lodash/get";

import { HelpButton } from "../Help";

import { ThumbnailSmall } from "./Thumbnail";
import {
  DocumentSearchFilter,
  FullTextSeachFilter,
} from "./SearchFilter";
import { LimitStringLength } from "../util";
import { useState } from "react";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import RadioButtonUncheckedIcon from "@mui/icons-material/RadioButtonUnchecked";
import ClearIcon from "@mui/icons-material/Clear";
import EditIcon from "@mui/icons-material/Edit";
import { Link } from "react-router-dom";

const cardStyle = {
  width: 280,
  minHeight: 400,
  margin: "0.5em",
  display: "inline-block",
  verticalAlign: "top",
};

const DocumentPagination = () => (
  <Pagination rowsPerPageOptions={[10, 25, 50, 100]} />
);

const DocumentListActions = () => (
  <TopToolbar>
    <DocumentHelp/>
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
        margin: "1em",
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
      <Typography variant="h6" color="textPrimary">
        Full text search
      </Typography>
      <p>
        Full text input filters documents on any fields available. For
        reference, see &nbsp;
        <a href="https://docs.meilisearch.com/learn/what_is_meilisearch/features.html">
          Meilisearch documentation
        </a>
      </p>
      <Typography variant="h6" color="textPrimary">
        Metadata filter
      </Typography>
      Returned documents can be filtered by their metadata. Possible queries:
      <Typography>- class:report</Typography>
      <Typography>- author:apple OR author:google</Typography>
      <Typography>
        - class:book AND (author:"agatha christie" OR author:"doyle")
      </Typography>
    </HelpButton>
  );
};

export const DocumentCard = (props: any) => {
  const { record } = props;
  const { selected, setSelected } = props;

  const isSelected = selected ? selected(record.id) : false;
  const select = () => (setSelected ? setSelected(record.id) : null);

  return (
    <Card
      key={record.id}
      style={cardStyle}
      sx={{
        borderRadius: "1em",
      }}
    >
      <CardHeader
        title={
          <Typography component="span" variant="body2">
            <span
              className="document-title"
              dangerouslySetInnerHTML={{
                __html: record ? LimitStringLength(record.name, 50) : "",
              }}
            />
          </Typography>
        }
        subheader={
          <Box display={{ xs: "block", sm: "flex" }} sx={{}}>
            <DateField
              record={record}
              source="date"
              style={{ textAlign: "left" }}
            />
            <Typography
              component="span"
              variant="body2"
              style={{ marginLeft: "11em", textAlign: "right" }}
            >
              {get(record, "type") ? get(record, "type") : ""}
            </Typography>
          </Box>
        }
        style={{
          flex: 1,
          height: "4em",
          background: "contrast",
          borderRadius: "15px",
        }}
      />
      <CardContent>
        <ThumbnailSmall url={record.preview_url} label="Img" />
      </CardContent>
      {
        <CardActions style={{ textAlign: "right" }}>
          <ShowButton resource="documents" record={record} />
          <EditButton resource="documents" record={record} />
          <ToggleButton
            size="small"
            value={record.id}
            selected={isSelected}
            onChange={select}
            sx={{
              borderWidth: "0px",
              background: "primary",
            }}
          >
            {isSelected ? (
              <CheckCircleIcon color="primary" />
            ) : (
              <RadioButtonUncheckedIcon />
            )}
          </ToggleButton>
        </CardActions>
      }
    </Card>
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
