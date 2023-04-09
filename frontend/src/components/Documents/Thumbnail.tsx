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
import { Button, useRecordContext } from "react-admin";
import { Paper, useMediaQuery, useTheme } from "@mui/material";
import DownloadIcon from "@mui/icons-material/Download";

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
  const theme = useTheme();
  const record = useRecordContext();
  const url = get(record, props.source || "");
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

  return record ? (
    <div
      style={{
        overflow: "hidden",
        maxHeight: "600px",
        minHeight: "500px",
        minWidth: "300px",
        maxWidth: "600px",
      }}
    >
      <img
        src={imgData}
        alt={props.label}
        style={{
          maxWidth: isSmall ? "350px" : isMedium ? "450px" : "600px",
          background: "white",
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
      <img
        src={imgData}
        style={{ maxWidth: "200px", background: "white" }}
        alt={props.label}
      />
    </div>
  );
}

export function EmbedFile({ source = "", filename = "" }) {
  const style = {
    width: "100%",
    display: "fill",
    border: "none",
    height: "100%",
  };

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
        });
      })
      .catch((response) => {
        console.error(response);
      });
  }, [record]);

  if (!record) return null;

  return (
    <>
      <a href={imgData} download={filename}>
        <Button color="primary" label={"Download"}>
          <DownloadIcon />
        </Button>
      </a>
      <Paper
        sx={{
          width: "100%",
          margin: "0.5em",
          height: "80vh",
        }}
      >
        <iframe style={style} title="Preview" src={imgData} />
      </Paper>
    </>
  );
}
