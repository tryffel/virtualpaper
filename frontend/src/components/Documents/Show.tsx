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
import HistoryIcon from "@mui/icons-material/History";
import { requestDocumentProcessing } from "../../api/dataProvider";
import { ThumbnailField, EmbedFile } from "./Thumbnail";
import { IndexingStatusField } from "./IndexingStatus";
import { MarkdownField } from "../Markdown";
import { number } from "prop-types";
import { PrettifyTime } from "../util";
import { ShowDocumentsEditHistory } from "./DocumentHistory";

export const DocumentShow = () => {
  const [enableFormatting, setState] = React.useState(true);
  const [historyEnabled, toggleHistory] = React.useState(false);
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
      actions={
        <DocumentShowActions
          showHistory={toggleHistory}
          historyShown={historyEnabled}
        />
      }
      aside={historyEnabled ? <ShowDocumentsEditHistory /> : undefined}
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
                <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                  <Typography variant="body2">Id: </Typography>
                  <TextField source="id" label="" variant="caption" />
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
        <Tab label="Processing trace">
          <DocumentJobsHistory />
        </Tab>
      </TabbedShowLayout>
    </Show>
  );
};

interface ActionsProps {
  historyShown: boolean;
  showHistory: (shown: boolean) => any;
}

function DocumentShowActions(props: ActionsProps) {
  const record = useRecordContext();
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));
  const requestProcessing = () => {
    if (record) {
      // @ts-ignore
      requestDocumentProcessing(record.id);
    }
  };

  const { historyShown, showHistory } = props;
  const toggleHistory = () => {
    showHistory(!historyShown);
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
      {!isSmall && (
        <Button
          color="primary"
          onClick={toggleHistory}
          label={historyShown ? "Hide history" : "Show history"}
        >
          <HistoryIcon />
        </Button>
      )}
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

export default DocumentShow;
