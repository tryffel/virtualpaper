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
  useRecordContext,
  useGetManyReference,
  Loading,
  Button,
} from "react-admin";
import {
  useMediaQuery,
  Box,
  Card,
  CardContent,
  Grid,
  Stepper,
  Step,
  StepLabel,
  StepContent,
  Typography,
} from "@mui/material";
import { PrettifyTime } from "../util";

import AddCircleIcon from "@mui/icons-material/AddCircle";
import UpdateIcon from "@mui/icons-material/Update";
import ArticleIcon from "@mui/icons-material/Article";
import ScheduleIcon from '@mui/icons-material/Schedule';

export const ShowDocumentsEditHistory = () => {
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
                data?.map((item: DocumentHistoryItem) => (
                  <ShowDocumentsEditHistoryItem item={item} />
                ))}
            </Stepper>
          </Grid>
        </CardContent>
      </Card>
    </Box>
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

const DocumentHistoryCreate = (props: HistoryProps) => {
  const { item, pretty_time } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<AddCircleIcon />}>Created document</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">Name: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDescription = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon/>}>Changed description</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">From: {item.old_value}</Typography>
        <Typography variant="body1">To: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRename = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon/>}>Renamed document</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">From: {item.old_value}</Typography>
        <Typography variant="body1">To: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryAddMetadata = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<UpdateIcon />}>Added metadata</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">Metadata: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRemoveMetadata = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<UpdateIcon />}>Added metadata</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">Metadata: {item.new_value}</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryContent = (props: HistoryProps) => {
  const { item, pretty_time } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon/>}>Changed content</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">Content</Typography>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDate = (props: HistoryProps) => {
  const { item, pretty_time } = props;

  const oldDate = new Date(item.old_value).toLocaleString();
  const newDate = new Date(item.new_value).toLocaleString();

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ScheduleIcon/>}>Changed date</StepLabel>
      <StepContent>
        <Typography variant="body2" gutterBottom>
          {item.user} - {pretty_time}:
        </Typography>
        <Typography variant="body1">From: {oldDate}</Typography>
        <Typography variant="body1">To: {newDate}</Typography>
      </StepContent>
    </Step>
  );
};
