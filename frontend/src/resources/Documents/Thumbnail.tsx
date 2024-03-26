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
import get from "lodash/get";
import { useNotify, useRecordContext, useTheme } from "react-admin";
import {
  CircularProgress,
  Paper,
  Typography,
  useMediaQuery,
} from "@mui/material";
import DownloadIcon from "@mui/icons-material/Download";
import MenuItem from "@mui/material/MenuItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";

export function downloadFile(url: string) {
  const token = localStorage.getItem("auth");
  return fetch(url, {
    method: "GET",
    headers: { Authorization: `Bearer ${token}` },
  });
}

export interface ThumbnailProps {
  source?: string;
  label: string;
  url?: string;
}

export const ThumbnailField = (props: any) => {
  const [theme] = useTheme();
  const record = useRecordContext();
  const [imgData, setImage] = React.useState("");
  const isMedium = useMediaQuery((theme: any) => theme.breakpoints.down("md"));
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  React.useEffect(() => {
    const url = get(record, props.source || "");
    downloadFile(url)
      .then((response) => {
        response.arrayBuffer().then((buffer) => {
          const data = window.URL.createObjectURL(new Blob([buffer]));
          setImage(data);
        });
      })
      .catch((response) => {
        console.error(response);
      });
  }, [record]);
  if (!record) return null;

  //sx={{ backgroundColor: theme === "dark" ? "#292929" : "#FAFAFA" }}

  return record ? (
    <div
      style={{
        overflow: "hidden",
        maxHeight: "600px",
        minHeight: "200px",
        minWidth: "300px",
        maxWidth: "450px",
      }}
    >
      <img
        src={imgData}
        alt={props.label}
        style={{
          backgroundColor: theme === "dark" ? "#D9D9D9" : "#FAFAFA",
          maxWidth: isSmall ? "280px" : isMedium ? "450px" : "450px",
          borderRadius: "5%",
          borderWidth: "thin",
          borderStyle: "solid",
          borderColor: "#ECEFF1",
        }}
      />
    </div>
  ) : null;
};

export function ThumbnailSmall(props: ThumbnailProps) {
  const [imgData, setImage] = React.useState(() => {
    downloadFile(props.url || "")
      .then((response) => {
        response.arrayBuffer().then((buffer) => {
          const data = window.URL.createObjectURL(new Blob([buffer]));
          setImage(data);
        });
      })
      .catch((response) => {
        console.error(response);
      });
    return "";
  });

  return (
    <div
      style={{
        overflow: "hidden",
        maxHeight: "200px",
        minHeight: "200px",
        minWidth: "150px",
        maxWidth: "200px",
        borderRadius: "5%",
      }}
    >
      <img src={imgData} style={{ maxWidth: "200px" }} alt={props.label} />
    </div>
  );
}

export function EmbedFile(
  props: {
    source: string;
    filename: string;
    setDownloadUrl?: (url: string) => void;
  } = { source: "", filename: "" }
) {
  const style = {
    width: "100%",
    display: "fill",
    border: "none",
    height: "100%",
  };

  const { source, filename, setDownloadUrl } = props;

  const record = useRecordContext();
  const url = get(record, source || "");
  const [imgData, setImage] = React.useState("");

  React.useEffect(() => {
    const url = get(record, source || "");
    downloadFile(url)
      .then((response) => {
        response.arrayBuffer().then((buffer) => {
          const data = window.URL.createObjectURL(new Blob([buffer]));
          setImage(data);
          setDownloadUrl && setDownloadUrl(data);
        });
      })
      .catch((response) => {
        console.error(response);
      });
  }, [record]);

  if (!record) return null;

  return (
    <Paper
      sx={{
        width: "100%",
        margin: "0.5em",
        height: "80vh",
      }}
    >
      <iframe style={style} title="Preview" src={imgData} />
    </Paper>
  );
}

export const DownloadDocumentButton = (props: { onFinished: () => void }) => {
  const [downloadClicked, setDownloadClicked] = React.useState(false);
  const handleClick = () => {
    if (downloadClicked) {
      return;
    }
    setDownloadClicked(true);
  };

  return (
    <MenuItem color={"primary"} onClick={handleClick}>
      <ListItemIcon>
        <DownloadIcon color={"primary"} />
      </ListItemIcon>
      <ListItemText>
        {downloadClicked && <DownloadFileLink onFinished={props.onFinished} />}
        <Typography variant="body1" color={"primary"}>
          Download
        </Typography>
      </ListItemText>
    </MenuItem>
  );
};

const DownloadFileLink = (props: { onFinished: () => void }) => {
  const [clicked, setClicked] = React.useState(false);
  const notify = useNotify();
  const record = useRecordContext();
  const [fileData, setFileData] = React.useState("");

  const url = get(record, "download_url");
  const filename = get(record, "filename");

  React.useEffect(() => {
    downloadFile(url)
      .then((response) => {
        response.arrayBuffer().then((buffer) => {
          const data = window.URL.createObjectURL(new Blob([buffer]));
          setFileData(data);
          notify("Document downloaded", { type: "info" });
        });
      })
      .catch((response) => {
        console.error(response);
        notify(`Error ${response.status}: ${response.error}`, {
          type: "error",
        });
      });
  }, []);

  React.useEffect(() => {
    if (fileData && !clicked) {
      setClicked(true);
      const link = document.createElement("a");
      link.href = fileData;
      link.setAttribute("download", filename);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      props.onFinished();
    }
  }, [fileData]);
  if (!fileData) {
    return <CircularProgress size={20} />;
  }

  return null;
};
