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
  Button,
  EditButton,
  Labeled,
  ShowBase,
  Tab,
  TabbedShowLayout,
  TextField,
  TopToolbar,
  useRecordContext,
} from "react-admin";
import {
  Box,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Grid,
  Paper,
  Typography,
  useMediaQuery,
} from "@mui/material";
import HistoryIcon from "@mui/icons-material/History";
import TimelineIcon from "@mui/icons-material/Timeline";
import CloseIcon from "@mui/icons-material/Close";
import {
  DownloadDocumentButton,
  EmbedFile,
  ThumbnailField,
} from "@components/document/fields/Thumbnail.tsx";
import { IndexingStatusField } from "@components/document/fields/IndexingStatus.tsx";
import { MarkdownField } from "@components/markdown";
import { ShowDocumentsEditHistory } from "@components/document/ShowHistory.tsx";
import { LinkedDocumentList } from "@components/document/fields/LinkedDocuments.tsx";
import { RequestIndexingModal } from "@components/document/edit/RequestIndexing.tsx";
import get from "lodash/get";
import MenuItem from "@mui/material/MenuItem";
import Menu from "@mui/material/Menu";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import NotesIcon from "@mui/icons-material/Notes";
import SourceIcon from "@mui/icons-material/Source";
import StickyNote2Icon from "@mui/icons-material/StickyNote2";
import { MetadataList } from "@components/document/fields/MetadataList.tsx";
import { SharedUsersList } from "@components/document/fields/SharedUsers";
import { EditDocumentSharing } from "@components/document/edit/EditSharing.tsx";
import { ShowDocumentContent } from "@components/document/ShowContent.tsx";
import { DocumentIdField } from "@components/document/fields/DocumentId.tsx";
import { TimestampField } from "@components/primitives/TimestampField.tsx";
import { DocumentJobsHistory } from "@components/document/fields/ProcessingHistory.tsx";
import {
  DocumentBasicInfo,
  DocumentTitle,
  DocumentTopRow,
} from "@components/document/fields/BasicInfo.tsx";
import { DocumentPropertyList } from "@components/properties/DocumentPropertyList.tsx";

export const DocumentShow = () => {
  const [asideMode, setAsideMode] = React.useState<AsideMode>("closed");
  const [downloadUrl, setDownloadUrl] = React.useState("");
  // @ts-ignore
  const isNotSmall = useMediaQuery((theme) => theme.breakpoints.up("sm"));
  // @ts-ignore
  const isNotMedium = useMediaQuery((theme) => theme.breakpoints.up("md"));
  const iconPosition = isNotSmall ? "start" : "top";

  const tabbedContentOverride = isNotSmall
    ? {}
    : { ".RaTabbedShowLayout-content": { pl: 0, pr: 0 } };

  return (
    <ShowBase>
      <>
        <Container sx={{ p: 0 }} maxWidth={"lg"}>
          <DocumentShowActions
            showHistory={() => setAsideMode("history")}
            showJobs={() => setAsideMode("jobs")}
            downloadUrl={downloadUrl}
          />
          <Box
            sx={{
              display: "flex",
              flexDirection: "row",
              gap: "10px",
              width: "100%",
            }}
          >
            <Paper elevation={3} sx={{ pl: 1, pr: 1, pt: 0.2, width: "100%" }}>
              <TabbedShowLayout
                sx={{
                  ".MuiTab-root": { minHeight: "36px" },
                  ...tabbedContentOverride,
                  marginBottom: 1,
                  marginTop: 1,
                }}
              >
                <Tab
                  label="general"
                  icon={<StickyNote2Icon />}
                  iconPosition={iconPosition}
                >
                  {isNotMedium ? (
                    <DocumentGeneralTabLarge />
                  ) : (
                    <DocumentGeneralTablSmall />
                  )}
                </Tab>
                <Tab
                  label="Content"
                  icon={<NotesIcon />}
                  iconPosition={iconPosition}
                >
                  <DocumentContentTab />
                </Tab>
                <Tab
                  label="preview"
                  icon={<SourceIcon />}
                  iconPosition={iconPosition}
                >
                  <DocumentPreviewTab setDownloadUrl={setDownloadUrl} />
                </Tab>
                <Box sx={{ ml: "auto", mr: 0 }}>
                  <IndexingStatusField source="status" />
                </Box>
              </TabbedShowLayout>
            </Paper>
          </Box>
        </Container>
        <DocumentShowAsideModal mode={asideMode} setMode={setAsideMode} />
      </>
    </ShowBase>
  );
};

interface ActionsProps {
  showHistory: (shown: boolean) => void;
  downloadUrl?: string;
  showJobs: (shown: boolean) => void;
}

function DocumentShowActions(props: ActionsProps) {
  const { showHistory, showJobs } = props;
  const toggleHistory = () => {
    showHistory(true);
    handleCloseMenu();
  };
  const toggleJobs = () => {
    showJobs(true);
    handleCloseMenu();
  };

  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const handleClickMenu = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleCloseMenu = () => {
    setAnchorEl(null);
  };

  const record = useRecordContext();
  if (!record) {
    return null;
  }

  return (
    <TopToolbar>
      <EditButton />
      <Button onClick={handleClickMenu} label={"More"}>
        <MoreVertIcon />
      </Button>
      <Menu anchorEl={anchorEl} open={!!anchorEl} onClose={handleCloseMenu}>
        <EditDocumentSharing onClose={handleCloseMenu} />
        <RequestIndexingModal onClose={handleCloseMenu} />
        <MenuItem color={"primary"} onClick={toggleHistory}>
          <ListItemIcon>
            <HistoryIcon color={"primary"} />
          </ListItemIcon>
          <ListItemText>
            <Typography variant="body1" color={"primary"}>
              Document History
            </Typography>
          </ListItemText>
        </MenuItem>
        <MenuItem color={"primary"} onClick={toggleJobs}>
          <ListItemIcon>
            <TimelineIcon color={"primary"} />
          </ListItemIcon>
          <ListItemText>
            <Typography variant="body1" color={"primary"}>
              Processing history
            </Typography>
          </ListItemText>
        </MenuItem>
        <DownloadDocumentButton onFinished={handleCloseMenu} />
      </Menu>
    </TopToolbar>
  );
}

export default DocumentShow;

const DocumentGeneralTabLarge = () => {
  return (
    <>
      <Grid container>
        <Grid item sm={9} xl={11} height={"fitContent"}>
          <DocumentTitle />
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid
          container
          item
          sm={6}
          md={8}
          xl={6}
          spacing={1}
          alignContent={"flex-start"}
        >
          <Grid item xs={12}>
            <DocumentBasicInfo />
          </Grid>
          <Grid item xs={12}>
            <ThumbnailField source="preview_url" label="Thumbnail" />
          </Grid>
          <Grid item xs={12}>
            <DocumentIdField />
          </Grid>

          <Grid item xs={12}></Grid>
        </Grid>

        <Grid
          container
          item
          sm={6}
          md={4}
          xl={4}
          spacing={3}
          alignContent={"space-between"}
        >
          <Grid item xs={12}>
            <MarkdownField source="description" />
          </Grid>
          <Grid item xs={12} sm={8}>
            <MetadataList />
          </Grid>
          <Grid item xs={12} sm={8}>
            <DocumentPropertyList />
          </Grid>
          <Grid item xs={12} sm={12}>
            <LinkedDocumentList />
          </Grid>
          <Grid item xs={12} sm={12}>
            <SharedUsersList />
          </Grid>
          <Grid item xs={12}>
            <Labeled label={"File size"}>
              <TextField source={"pretty_size"} />
            </Labeled>
          </Grid>
          <Grid item xs={12}>
            <TimestampField />
          </Grid>
        </Grid>
      </Grid>
    </>
  );
};

const DocumentGeneralTablSmall = () => {
  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <DocumentTopRow />
      </Grid>
      <Grid item xs={12}>
        <ThumbnailField source="preview_url" label="Thumbnail" />
      </Grid>
      <Grid item xs={12}>
        <Labeled label="Description">
          <MarkdownField source="description" />
        </Labeled>
      </Grid>
      <Grid item xs={12}>
        <MetadataList />
      </Grid>
      <Grid item xs={12}>
        <DocumentPropertyList />
      </Grid>
      <Grid item xs={12}>
        <LinkedDocumentList />
      </Grid>
      <Grid item xs={12} sm={6}>
        <Labeled label={"File size"}>
          <TextField source={"pretty_size"} />
        </Labeled>
      </Grid>
      <Grid item xs={12} sm={6}>
        <SharedUsersList />
      </Grid>
      <Grid item xs={12}>
        <TimestampField />
      </Grid>
    </Grid>
  );
};

const DocumentContentTab = () => {
  return <ShowDocumentContent />;
};

const DocumentPreviewTab = (props: {
  setDownloadUrl: (url: string) => void;
}) => {
  const record = useRecordContext();
  return (
    <EmbedFile
      source="download_url"
      filename={get(record, "filename")}
      {...props}
    />
  );
};

type AsideMode = "closed" | "history" | "jobs";

interface AsideProps {
  mode: AsideMode;
  setMode: (mode: AsideMode) => void;
}

const DocumentShowAsideModal = (props: AsideProps) => {
  const { mode, setMode } = props;

  const title = mode === "history" ? "Document history" : "Processing history";

  return (
    <Dialog open={mode !== "closed"} scroll="paper">
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        {mode == "history" && <ShowDocumentsEditHistory />}
        {mode == "jobs" && <DocumentJobsHistory />}
      </DialogContent>
      <DialogActions>
        <Button label={"Close"} onClick={() => setMode("closed")}>
          <CloseIcon />
        </Button>
      </DialogActions>
    </Dialog>
  );
};
