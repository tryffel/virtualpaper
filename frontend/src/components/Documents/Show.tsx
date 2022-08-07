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

import React, { useState } from "react";
import {
  Show,
  TabbedShowLayout,
  Tab,
  TextField,
  DateField,
  Labeled,
  ArrayField,
  SingleFieldList,
  ChipField,
  Datagrid,
  Button,
  TopToolbar,
  EditButton,
  Loading,
  useGetManyReference,
  useRecordContext,
  useGetList,
} from "react-admin";
import {
  Accordion,
  Typography,
  AccordionSummary,
  AccordionDetails,
  Grid,
  Box,
  Card,
  CardContent,
  Stepper,
  Step,
  StepLabel,
  StepContent,
  useMediaQuery,
} from "@mui/material";
import { Repeat, ExpandMore } from "@mui/icons-material";
import { requestDocumentProcessing } from "../../api/dataProvider";
import { ThumbnailField, EmbedFile } from "./Thumbnail";
import { IndexingStatusField } from "./IndexingStatus";
import { MarkdownField } from "../Markdown";

export const DocumentShow = () => {
  const [enableFormatting, setState] = React.useState(true);
  const record = useRecordContext();

  const toggleFormatting = () => {
    if (enableFormatting) {
      setState(false);
    } else {
      setState(true);
    }
  };

  return (
    <Show
      title="Document"
      actions={<DocumentShowActions />}
      aside={<ShowDocumentsEditHistory />}
    >
      <TabbedShowLayout>
        <Tab label="general">
          <div>
            <Grid container spacing={2}>
              <Grid item xs={12} md={8} lg={12}>
                <Box display={{ xs: "block", sm: "flex" }}>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <Typography>Name</Typography>
                    <TextField
                      source="name"
                      label=""
                      style={{ fontSize: "2em" }}
                    />
                  </Box>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <Typography>Date</Typography>
                    <DateField source="date" showTime={false} label="Date" />
                  </Box>
                </Box>
                <IndexingStatusField source="status" label="" />
                <Box display={{ xs: "block", sm: "flex" }}>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <ThumbnailField source="preview_url" label="Thumbnail" />
                  </Box>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <Labeled label="Description">
                      <MarkdownField source="description" />
                    </Labeled>
                  </Box>
                  {record && record.tags ? (
                    <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                      <ArrayField source="tags">
                        <SingleFieldList>
                          <ChipField source="key" />
                        </SingleFieldList>
                      </ArrayField>
                    </Box>
                  ) : null}
                </Box>
                <Box display={{ xs: "block", sm: "flex" }}>
                  <Typography>Type</Typography>
                  <ChipField source="type"></ChipField>
                </Box>
                <Box display={{ xs: "block", sm: "flex" }}>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <Typography>Metadata</Typography>
                    <ArrayField source="metadata">
                      <Datagrid bulkActionButtons={false}>
                        <TextField source="key" />
                        <TextField source="value" />
                      </Datagrid>
                    </ArrayField>
                  </Box>
                </Box>
                <Box display={{ xs: "block", sm: "flex" }}>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <Typography>Uploaded at</Typography>
                    <DateField
                      source="created_at"
                      label="Uploaded"
                      showTime={false}
                    />
                  </Box>
                  <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                    <Typography>Updated at</Typography>
                    <DateField
                      source="updated_at"
                      label="Last updated"
                      showTime
                    />
                  </Box>
                </Box>
              </Grid>
            </Grid>
          </div>
        </Tab>
        <Tab label="Content">
          <Button
            color="primary"
            size="medium"
            variant="contained"
            onClick={toggleFormatting}
          >
            <Typography variant="h6">
              {enableFormatting ? "Enable formatting" : "Disable formatting"}
            </Typography>
          </Button>
          {enableFormatting ? (
            <TextField source="content" label="Raw parsed text content" />
          ) : (
            <MarkdownField source="content" label="Raw parsed text content" />
          )}
        </Tab>
        <Tab label="preview">
          <EmbedFile source="download_url" />
        </Tab>
        <Tab label="history">
          <DocumentJobsHistory />
        </Tab>
      </TabbedShowLayout>
    </Show>
  );
};

function DocumentShowActions() {
  const record = useRecordContext();
  const requestProcessing = () => {
    if (record) {
      // @ts-ignore
      requestDocumentProcessing(record.id);
    }
  };

  return (
    <TopToolbar>
      <EditButton />
      <Button
        color="primary"
        onClick={requestProcessing}
        label="Request re-processing"
      >
        <Repeat />
      </Button>
    </TopToolbar>
  );
}

function DocumentJobsHistory() {
  const record = useRecordContext();
  const { data, isLoading, error } = useGetManyReference("document/jobs", {
    target: "id",
    id: record?.id,
    sort: {
      field: "timestamp",
      order: "ASC",
    },
  });

  if (isLoading) {
    return <Loading />;
  }
  if (error) {
    return null;
    // return <Error />;
  }

  if (data !== undefined) {
    return (
      <div>
        {data.map((index) => (
          <DocumentJobListItem record={index} />
        ))}
      </div>
    );
  }
  return null;
}

function DocumentJobListItem(props: any) {
  if (!props.record) {
    return null;
  }
  const ok = props.record.status === "Finished";
  let style = {};
  let prefix = "";
  if (props.record.status === "Finished") {
  } else if (props.record.status === "Running") {
    style = { fontStyle: "italic", background: "#ff0" };
    prefix = "Running";
  } else if (props.record.status === "Error") {
    style = { fontStyle: "italic", background: "red" };
    prefix = "Error";
  }

  return (
    <Accordion>
      <AccordionSummary expandIcon={<ExpandMore />} style={style}>
        <Typography>
          {prefix} {props.record.message}
        </Typography>
      </AccordionSummary>
      <AccordionDetails style={{ flexDirection: "column" }}>
        <Typography>
          Status:
          {props.record.status}
        </Typography>
        <Typography>
          Job id:
          {props.record.id}
        </Typography>
        <Typography>
          Started at:
          {props.record.started_at}
        </Typography>
        <Typography>
          Stopped at:
          {props.record.stopped_at}
        </Typography>
      </AccordionDetails>
    </Accordion>
  );
}

const ShowDocumentsEditHistory = () => {
  const [shown, setShown] = useState(false);

  const record = useRecordContext();

  const { data, isLoading, error } = useGetManyReference(
    "documents/edithistory",
    {
      target: "id",
      id: record?.id,
      sort: {
        field: "created_at",
        order: "DESC",
      },
    }
  );

  const toggle = () => {
    setShown(!shown);
  };

  const isMd = useMediaQuery((theme: any) => theme.breakpoints.down("md"));
  if (isMd) {
    return null;
  }

  if (isLoading) {
    return <Loading />;
  }
  if (error) {
    return null;
  }

  return (
    <Box ml={2}>
      <Card>
        <CardContent>
          <Grid container flex={1}>
            <Grid item xs={12} md={6}>
              <Box flexGrow={0}>
                <Button label="Toggle history" onClick={toggle} />
              </Box>
            </Grid>

            <Stepper orientation="vertical" sx={{ mt: 1 }}>
              {shown &&
                data?.map((item: any) => (
                  <Step key={`${item.id}`} expanded active completed>
                    <StepContent>
                      <Typography variant="body2" gutterBottom>
                        {item.created_at}:
                      </Typography>
                      <Typography variant="body1">{item.action}</Typography>
                      <Typography variant="body1">
                        From: {item.old_value}
                      </Typography>
                      <Typography variant="body1">
                        To: {item.new_value}
                      </Typography>
                    </StepContent>
                  </Step>
                ))}
            </Stepper>
          </Grid>
        </CardContent>
      </Card>
    </Box>
  );
};

export default DocumentShow;
