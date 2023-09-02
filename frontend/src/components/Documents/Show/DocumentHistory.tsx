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
import { useRecordContext, useGetManyReference, Loading } from "react-admin";
import {
  Grid,
  Stepper,
  Step,
  StepLabel,
  StepContent,
  Typography,
  Tooltip,
} from "@mui/material";
import { PrettifyTime } from "../../util";

import AddCircleIcon from "@mui/icons-material/AddCircle";
import ArticleIcon from "@mui/icons-material/Article";
import ScheduleIcon from "@mui/icons-material/Schedule";
import TagIcon from "@mui/icons-material/Tag";
import ContentCut from "@mui/icons-material/ContentCut";
import FormatListBulletedIcon from "@mui/icons-material/FormatListBulleted";
import DeleteIcon from "@mui/icons-material/Delete";
import RestoreFromTrashIcon from "@mui/icons-material/RestoreFromTrash";
import TranslateIcon from "@mui/icons-material/Translate";

import get from "lodash/get";
import { languages } from "../../../languages";

export const ShowDocumentsEditHistory = () => {
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

  if (isLoading) {
    return <Loading />;
  }
  if (error) {
    return null;
  }

  return (
    <Grid container>
      <Grid item xs={8} md={8}>
        <Typography variant="overline">Document edit history</Typography>
      </Grid>
      <Grid item xs={12} md={12}>
        <Stepper orientation="vertical" sx={{ mt: 1 }}>
          {data?.map((item: DocumentHistoryItem) => (
            <ShowDocumentsEditHistoryItem item={item} />
          ))}
        </Stepper>
      </Grid>
    </Grid>
  );
};

interface DocumentHistoryItem {
  id: number;
  document_id: string;
  action: string;
  old_value: string;
  new_value: string;
  user_id: number;
  user: number;
  created_at: string | number;
}

interface HistoryProps {
  item: DocumentHistoryItem;
  pretty_time: string;
}

const ShowDocumentsEditHistoryItem = (props: { item: DocumentHistoryItem }) => {
  const { item } = props;
  const timeString = PrettifyTime(item.created_at);

  switch (item.action) {
    case "create":
      return <DocumentHistoryCreate pretty_time={timeString} item={item} />;
    case "rename":
      return <DocumentHistoryRename pretty_time={timeString} item={item} />;
    case "content":
      return <DocumentHistoryContent pretty_time={timeString} item={item} />;
    case "description":
      return (
        <DocumentHistoryDescription pretty_time={timeString} item={item} />
      );
    case "date":
      return <DocumentHistoryDate pretty_time={timeString} item={item} />;
    case "add metadata":
      return (
        <DocumentHistoryAddMetadata pretty_time={timeString} item={item} />
      );
    case "remove metadata":
      return (
        <DocumentHistoryRemoveMetadata pretty_time={timeString} item={item} />
      );
    case "delete":
      return <DocumentHistoryDelete pretty_time={timeString} item={item} />;
    case "restore":
      return <DocumentHistoryRestore pretty_time={timeString} item={item} />;
    case "lang":
      return <DocumentHistoryLang item={item} pretty_time={timeString} />;
    case "modified linked documents":
      return (
        <DocumentHistoryModifyLinkedDocuments
          pretty_time={timeString}
          item={item}
        />
      );
  }

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel>label: {item.action}</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {timeString}:
        </Typography>
        <Typography variant="body1">{item.action}</Typography>
        <Typography variant="body1">From: {item.old_value}</Typography>
        <Typography variant="body1">To: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

// create, rename, add metadata, remove metadata, date, description, content
//
const ItemLabel = (props: HistoryProps) => {
  const ref = React.createRef();
  const { item, pretty_time } = props;
  // @ts-ignore
  const fullTime = new Date(Date.parse(item.created_at)).toLocaleString();

  return (
    <Tooltip title={`Time: ${fullTime}`}>
      <Typography variant="body2" gutterBottom>
        {item.user} - {pretty_time}:
      </Typography>
    </Tooltip>
  );
};

const DocumentHistoryCreate = (props: HistoryProps) => {
  const { item, pretty_time } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<AddCircleIcon />}>Created document</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">Name: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDescription = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon />}>Changed description</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">To: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRename = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon />}>Renamed document</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">To: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryAddMetadata = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  let keyId = "";
  let valueId = "";
  let parsed = {};
  let jsonMode = true;
  try {
    parsed = JSON.parse(item.new_value);
    keyId = get(parsed, "key_id");
    valueId = get(parsed, "value_id");
  } catch (e) {
    // old mode without json and notation is key_name:value_name
    [keyId, valueId] = item.new_value.split(":");
    jsonMode = false;
  }

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<TagIcon />}>Added metadata</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        {jsonMode ? (
          <Tooltip
            title={`Metadata info: key_id: ${keyId}, value_id: ${valueId}`}
          >
            <Typography variant="body1">No data</Typography>
          </Tooltip>
        ) : (
          <Typography variant="body1">
            Metadata: {keyId}:{valueId}
          </Typography>
        )}
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRemoveMetadata = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  let keyId = "";
  let valueId = "";
  let parsed = {};
  let jsonMode = true;
  try {
    parsed = JSON.parse(item.old_value);
    keyId = get(parsed, "key_id");
    valueId = get(parsed, "value_id");
  } catch (e) {
    [keyId, valueId] = item.old_value.split(":");
    jsonMode = false;
  }
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ContentCut />}>Removed metadata</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        {jsonMode ? (
          <Tooltip
            title={`Metadata info: key_id: ${keyId}, value_id: ${valueId}`}
          >
            <Typography variant="body1">No data</Typography>
          </Tooltip>
        ) : (
          <Typography variant="body1">
            Metadata: {keyId}:{valueId}
          </Typography>
        )}
      </StepContent>
    </Step>
  );
};

const DocumentHistoryContent = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon />}>Changed content</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">Content</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDate = (props: HistoryProps) => {
  const { item, pretty_time } = props;

  const oldDate = new Date(Number(item.old_value) * 1000).toLocaleDateString();
  const newDate = new Date(Number(item.new_value) * 1000).toLocaleDateString();

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ScheduleIcon />}>Changed date</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">To: {newDate}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryModifyLinkedDocuments = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<FormatListBulletedIcon />}>
        Modified linked documents
      </StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">{item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDelete = (props: HistoryProps) => {
  const { item } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<DeleteIcon />}>Deleted</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRestore = (props: HistoryProps) => {
  const { item } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<RestoreFromTrashIcon />}>Restored</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryLang = (props: HistoryProps) => {
  const { item } = props;
  const newLang = languages[props.item.new_value as keyof typeof languages];

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<TranslateIcon />}>Set language</StepLabel>
      <StepContent>
        <ItemLabel {...props} />
        <Typography variant="body1">{newLang}</Typography>
      </StepContent>
    </Step>
  );
};
