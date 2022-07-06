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
  RichTextField,
  ShowButton,
  useListContext,
  Pagination,
  TopToolbar,
  SortButton,
  ExportButton,
  CreateButton,
  Button,
} from "react-admin";
import {
  useMediaQuery,
  Grid,
  Card,
  CardContent,
  CardActions,
  CardHeader,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Typography,
  Box,
  useTheme,
} from "@mui/material";
import get from "lodash/get";

import { Help } from "@mui/icons-material";

import { ThumbnailSmall } from "./Thumbnail";
import { DocumentSearchFilter, FilterSidebar } from "./SearchFilter";

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
    <HelpButton />
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
      filters={DocumentSearchFilter}
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
      aside={<FilterSidebar />}
      actions={<DocumentListActions />}
      sort={{ field: "date", order: "DESC" }}
      filters={DocumentSearchFilter}
    >
      <DocumentGrid />
    </List>
  );
};

const DocumentGrid = () => {
  const { data, isLoading } = useListContext();
  const theme = useTheme();

  return !isLoading && data ? (
    <Grid
      flex={2}
      sx={{
        background: theme.palette.background.default,
        margin: "1em",
      }}
    >
      {data.map((record) => (
        <Card
          key={record.id}
          style={cardStyle}
          sx={{
            borderRadius: "1em",
          }}
        >
          <CardHeader
            title={<RichTextField record={record} source="name" />}
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
            </CardActions>
          }
        </Card>
      ))}
    </Grid>
  ) : null;
};

const HelpDialog = (props: any) => {
  const [scroll, setScroll] = React.useState("paper");

  const { onClose, open } = props;
  const handleClose = () => {
    onClose();
  };

  return (
    <Dialog
      onClose={handleClose}
      aria-labelledby="simple-dialog-title"
      open={open}
    >
      <DialogTitle id="simple-dialog-title">
        Help: Querying documents
      </DialogTitle>
      <DialogContent dividers={scroll === "paper"}>
        <DialogContentText>
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
          Returned documents can be filtered by their metadata. Possible
          queries:
          <Typography>- class:report</Typography>
          <Typography>- author:apple OR author:google</Typography>
          <Typography>
            - class:book AND (author:"agatha christie" OR author:"doyle")
          </Typography>
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>
          <Typography>Close</Typography>
        </Button>
      </DialogActions>
    </Dialog>
  );
};

const HelpButton = () => {
  const [open, setOpen] = React.useState(false);

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  return (
    <div>
      <Button
        label="Help"
        size="small"
        alignIcon="left"
        onClick={handleClickOpen}
      >
        <Help />
      </Button>
      <HelpDialog open={open} onClose={handleClose} />
    </div>
  );
};
