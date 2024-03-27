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
  DateField,
  EditButton,
  Labeled,
  Show,
  Tab,
  TabbedShowLayout,
  TextField,
  TopToolbar,
  useRecordContext,
} from "react-admin";
import {
  Box,
  Card,
  CardContent,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Grid,
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
} from "../Thumbnail";
import { IndexingStatusField } from "../IndexingStatus";
import { MarkdownField } from "../../../components/markdown";
import { ShowDocumentsEditHistory } from "./DocumentHistory";
import { LinkedDocumentList } from "./LinkedDocuments";
import {
  DocumentBasicInfo,
  DocumentIdField,
  DocumentJobsHistory,
  DocumentTitle,
  DocumentTopRow,
} from "./Show";
import { RequestIndexingModal } from "../RequestIndexing";
import get from "lodash/get";
import MenuItem from "@mui/material/MenuItem";
import Menu from "@mui/material/Menu";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import NotesIcon from "@mui/icons-material/Notes";
import SourceIcon from "@mui/icons-material/Source";
import StickyNote2Icon from "@mui/icons-material/StickyNote2";
import { MetadataList } from "./MetadataList";
import { ListSharedUsers } from "./SharedUsers";
import { EditDocumentSharing } from "../EditSharing";
import { ShowDocumentContent } from "../../../components/document/ShowContent.tsx";

export const DocumentShow = () => {
  const [asideMode, setAsideMode] = React.useState<AsideMode>("closed");
  const [downloadUrl, setDownloadUrl] = React.useState("");
  const isNotSmall = useMediaQuery((theme: any) => theme.breakpoints.up("sm"));
  const isNotMedium = useMediaQuery((theme: any) => theme.breakpoints.up("md"));
  const iconPosition = isNotSmall ? "start" : "top";

  return (
    <Show
      title="Document"
      actions={
        <DocumentShowActions
          showHistory={() => setAsideMode("history")}
          showJobs={() => setAsideMode("jobs")}
          downloadUrl={downloadUrl}
        />
      }
      aside={
        isNotSmall ? (
          <DocumentShowAside mode={asideMode} setMode={setAsideMode} />
        ) : undefined
      }
    >
      <Container>
        <TabbedShowLayout
          sx={{
            ".MuiTab-root": { minHeight: "36px" },
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
          <Tab label="Content" icon={<NotesIcon />} iconPosition={iconPosition}>
            <DocumentContentTab />
          </Tab>
          <Tab
            label="preview"
            icon={<SourceIcon />}
            iconPosition={iconPosition}
          >
            <DocumentPreviewTab setDownloadUrl={setDownloadUrl} />
          </Tab>
        </TabbedShowLayout>
      </Container>
      {!isNotSmall && (
        <DocumentShowAsideModal mode={asideMode} setMode={setAsideMode} />
      )}
    </Show>
  );
};

interface ActionsProps {
  showHistory: (shown: boolean) => any;
  downloadUrl?: string;
  showJobs: (shown: boolean) => any;
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
        <Grid item sm={3} xl={1} alignContent={"flex-end"}>
          <IndexingStatusField source="status" label="" />
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
          <Grid item xs={12} sm={12}>
            <LinkedDocumentList />
          </Grid>
          <Grid item xs={12} sm={12}>
            <ListSharedUsers />
          </Grid>
          <Grid item xs={4} sm={2} md={4} alignContent={"flex-end"}>
            <Labeled label={"File size"}>
              <TextField source={"pretty_size"} />
            </Labeled>
          </Grid>
          <Grid item xs={4} sm={2} md={4}>
            <Labeled label="Uploaded">
              <DateField source="created_at" showTime={false} />
            </Labeled>
          </Grid>
          <Grid item xs={4} sm={2} md={4}>
            <Labeled label={"Last updated"}>
              <DateField source="updated_at" showTime />
            </Labeled>
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
        <IndexingStatusField source="status" label="" />
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
        <LinkedDocumentList />
      </Grid>
      <Grid item xs={12}>
        <ListSharedUsers />
      </Grid>
      <Grid item xs={4} sm={3}>
        <Labeled label={"File size"}>
          <TextField source={"pretty_size"} />
        </Labeled>
      </Grid>
      <Grid item xs={4} sm={3}>
        <Labeled label="Uploaded">
          <DateField source="created_at" showTime={false} />
        </Labeled>
      </Grid>
      <Grid item xs={4} sm={3}>
        <Labeled label={"Last updated"}>
          <DateField source="updated_at" showTime />
        </Labeled>
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

const DocumentShowAside = (props: AsideProps) => {
  const { mode, setMode } = props;

  if (mode == "closed") {
    return null;
  }

  return (
    <Box ml={1} sx={{ maxWidth: "30%" }}>
      <Card>
        <CardContent>
          <Button
            label={"Close"}
            variant="outlined"
            onClick={() => setMode("closed")}
            sx={{ mb: 2 }}
          />
          {mode == "history" && <ShowDocumentsEditHistory />}
          {mode == "jobs" && <DocumentJobsHistory />}
        </CardContent>
      </Card>
    </Box>
  );
};

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
