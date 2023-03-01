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

import React from "react";
import {
  ArrayField,
  Button,
  ChipField,
  Datagrid,
  DateField,
  EditButton,
  Labeled,
  Show,
  SingleFieldList,
  Tab,
  TabbedShowLayout,
  TextField,
  TopToolbar,
  useRecordContext,
} from "react-admin";
import { Box, Grid, Typography, useMediaQuery, Divider } from "@mui/material";
import { Repeat } from "@mui/icons-material";
import HistoryIcon from "@mui/icons-material/History";
import { requestDocumentProcessing } from "../../../api/dataProvider";
import { EmbedFile, ThumbnailField } from "../Thumbnail";
import { IndexingStatusField } from "../IndexingStatus";
import { MarkdownField } from "../../Markdown";
import { ShowDocumentsEditHistory } from "./DocumentHistory";
import { LinkedDocumentList } from "./LinkedDocuments";
import { DocumentJobsHistory, DocumentTopRow } from "./Show";
import { RequestIndexingModal } from "../RequestIndexing";

export const DocumentShow = () => {
  const [historyEnabled, toggleHistory] = React.useState(false);

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
          <DocumentGeneralTab />
        </Tab>
        <Tab label="Content">
          <DocumentContentTab />
        </Tab>
        <Tab label="preview">
          <DocumentPreviewTab />
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
  const { historyShown, showHistory } = props;
  const toggleHistory = () => {
    showHistory(!historyShown);
  };

  return (
    <TopToolbar>
      <EditButton />
      <RequestIndexingModal />
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

export default DocumentShow;

const DocumentGeneralTab = () => {
  const record = useRecordContext();

  return (
    <Grid container spacing={2}>
      <Grid item xs={12} md={8} lg={12}>
        <DocumentTopRow />
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
          <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
            <Labeled label="Linked documents">
              <LinkedDocumentList />
            </Labeled>
          </Box>
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
            <DateField source="created_at" label="Uploaded" showTime={false} />
          </Box>
          <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography>Updated at</Typography>
            <DateField source="updated_at" label="Last updated" showTime />
          </Box>
        </Box>
      </Grid>
    </Grid>
  );
};

const DocumentContentTab = () => {
  const [enableFormatting, setState] = React.useState(true);

  const toggleFormatting = () => {
    if (enableFormatting) {
      setState(false);
    } else {
      setState(true);
    }
  };

  return (
    <Grid container maxWidth={800}>
      <Grid item sx={{ pb: 3, pt: 2 }}>
        <Box
          style={{
            display: "flex",
            flexFlow: " row",
            justifyContent: "flex-end",
          }}
        >
          <Button
            color="primary"
            size="small"
            variant="contained"
            onClick={toggleFormatting}
            sx={{ mr: 4 }}
          >
            <div>
              {enableFormatting ? "Enable formatting" : "Disable formatting"}
            </div>
          </Button>
          <div style={{ maxWidth: 400 }}>
            <Typography variant="body2">
              This page show automatically extracted content for the document.
              The quality and accuracy may vary depending on document type and
              quality.
            </Typography>
          </div>
        </Box>
      </Grid>
      <Grid item>
        <Typography variant="h5">Document content</Typography>
        <Divider sx={{ pt: 1 }} />
        {enableFormatting ? (
          <TextField source="content" label="Raw parsed text content" />
        ) : (
          <MarkdownField source="content" label="Raw parsed text content" />
        )}
      </Grid>
    </Grid>
  );
};

const DocumentPreviewTab = () => {
  return <EmbedFile source="download_url" />;
};
