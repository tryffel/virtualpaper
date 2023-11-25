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
import {
  Create,
  Error,
  FileField,
  FileInput,
  HttpError,
  Loading,
  SaveButton,
  SimpleForm,
  Toolbar,
  useCreate,
  useGetOne,
  useNotify,
  useCreatePath,
  Button,
} from "react-admin";
import { useEffect } from "react";
import { Box, ListItemButton, Typography } from "@mui/material";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import AddCircleOutlineIcon from "@mui/icons-material/AddCircleOutline";
import ListItemIcon from "@mui/material/ListItemIcon";
import PriorityHighIcon from "@mui/icons-material/PriorityHigh";
import ClearIcon from "@mui/icons-material/Clear";
import { useFormContext } from "react-hook-form";
import { Link } from "react-router-dom";

interface UploadError extends HttpError {
  body: {
    error: string;
    id: string;
    name: string;
  };
}

type Link = {
  id: string;
  name: string;
  created: boolean;
};

export const DocumentCreate = () => {
  const {
    data,
    isLoading: fileTypesLoading,
    error,
  } = useGetOne("filetypes", { id: "" });
  const notify = useNotify();
  const [fileNames, setFileNames] = React.useState([]);
  const [mimeTypes, setMimeTypes] = React.useState([]);
  const [links, setLinks] = React.useState<Link[]>([]);
  const [uploadData, setUploadData] = React.useState([]);
  const [create] = useCreate("documents", undefined, {
    onSuccess: (data) => {
      setLinks(
        links?.concat({
          id: data.id,
          name: data.name,
          created: true,
        })
      );
    },
    onError: (data: UploadError) => {
      if (data.status === 400 && data.body.error === "document exists") {
        setLinks(
          links?.concat({
            id: data.body.id,
            name: data.body.name,
            created: false,
          })
        );
      } else {
        notify(`Error: ${data.status}`);
        console.error(data);
      }
    },
  });

  useEffect(() => {
    if (data) {
      setFileNames(data.names.join(", "));
      setMimeTypes(data.mimetypes.join(", "));
    }
  }, [data]);

  const handleSubmit = async (data: any) => {
    setLinks([]);
    setUploadData(data.file.map(() => 1));
    return Promise.all(
      data.file.map((file: any, index: number) => {
        setTimeout(() => create("documents", { data: { file } }), 300 * index);
      })
    )
      .then((data) => {
        notify(`${data.length} files uploaded`, { type: "info" });
      })
      .catch((err) => {
        notify(`Error`, { type: "error" });
        console.error("upload files", err);
      });
  };

  const handleReset = () => {
    setLinks([]);
    setUploadData([]);
  };

  if (fileTypesLoading) return <Loading />;

  if (error) {
    // @ts-ignore
    return <Error error={error} />;
  }

  return (
    // @ts-ignore
    <Create title="Upload documents">
      <SimpleForm
        title={"Upload new document"}
        toolbar={
          <Toolbar>
            <UploadButton />
            <ClearButton reset={handleReset} />
          </Toolbar>
        }
        onSubmit={handleSubmit}
      >
        <Typography variant="body2">
          Supported file types: <em className="mimetypes">{fileNames}</em>
        </Typography>
        <FileInput
          accept={mimeTypes}
          multiple={true}
          label="File upload"
          source="file"
        >
          <FileField source="src" title="title" />
        </FileInput>

        {links.length > 0 && (
          <Box mt={3}>
            <Typography mb={1}>
              Uploaded documents: {links.length} / {uploadData.length}
            </Typography>
            <List>
              {links?.map((link) => (
                <DocumentLink {...link} />
              ))}
            </List>
          </Box>
        )}
      </SimpleForm>
    </Create>
  );
};

const UploadButton = () => {
  return <SaveButton label={"Upload"} />;
};

const ClearButton = ({ reset: resetOuter }: { reset: () => void }) => {
  const { reset, formState } = useFormContext();

  const handleReset = () => {
    reset();
    resetOuter();
  };

  return (
    <Button
      label={"Clear"}
      startIcon={<ClearIcon />}
      onClick={handleReset}
      variant={"contained"}
      size={"medium"}
      sx={{ marginLeft: 1 }}
      disabled={!formState.isDirty}
    />
  );
};

const DocumentLink = (link: Link) => {
  const createPath2 = useCreatePath();
  const icon = link.created ? (
    <AddCircleOutlineIcon color={"info"} />
  ) : (
    <PriorityHighIcon color={"warning"} />
  );

  return (
    <ListItem>
      <ListItemButton
        href={
          `/#` +
          createPath2({
            resource: "documents",
            type: "show",
            id: link.id,
          })
        }
      >
        <ListItemIcon>{icon}</ListItemIcon>
        <ListItemText
          primary={<Typography variant={"body1"}>{link.name}</Typography>}
          secondary={
            !link.created && (
              <Typography variant={"body2"}>Existing document</Typography>
            )
          }
        />
      </ListItemButton>
    </ListItem>
  );
};
