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
  SimpleForm,
  useGetOne,
  useNotify,
  useRedirect,
} from "react-admin";
import { useEffect } from "react";
import { Typography } from "@mui/material";

interface UploadError extends HttpError {
  body: {
    error: string;
    id: string;
    name: string;
  };
}

export const DocumentCreate = () => {
  const notify = useNotify();
  const redirect = useRedirect();
  const { data, isLoading, error } = useGetOne("filetypes", { id: "" });

  const [fileNames, setFileNames] = React.useState("");
  const [mimeTypes, setMimeTypes] = React.useState("");

  useEffect(() => {
    if (data) {
      setFileNames(data.names.join(", "));
      setMimeTypes(data.mimetypes.join(", "));
    }
  }, [data]);

  const onError = (data: UploadError) => {
    if (data.status === 400 && data.body.error === "document exists") {
      notify("Duplicate document found. Showing existing document", {
        type: "info",
        autoHideDuration: 4000,
      });
      redirect("show", "documents", data.body.id);
    }
  };

  if (isLoading) return <Loading />;

  if (error) {
    // @ts-ignore
    return <Error error={error} />;
  }

  return (
    // @ts-ignore
    <Create mutationOptions={{ onError }} title="Upload document">
      <SimpleForm title={"Upload new document"}>
        <Typography variant="body2">
          Supported file types: <em className="mimetypes">{fileNames}</em>
        </Typography>
        <FileInput
          accept={mimeTypes}
          multiple={false}
          label="File upload"
          source="file"
        >
          <FileField source="src" title="title" />
        </FileInput>
      </SimpleForm>
    </Create>
  );
};
