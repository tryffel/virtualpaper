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

import {
  useRecordContext,
  useGetManyReference,
  Loading,
  RecordContextProvider,
} from "react-admin";
import {
  Stepper,
  Step,
  StepLabel,
  StepContent,
  Typography,
  Tooltip,
  Box,
} from "@mui/material";
import { prettifyRelativeTime } from "@components/util";

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
import { languages } from "@/languages";
import { MarkdownField } from "@components/markdown";
import BookmarkIcon from "@mui/icons-material/Bookmark";
import BookmarkBorderIcon from "@mui/icons-material/BookmarkBorder";

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
    },
  );

  if (isLoading) {
    return <Loading />;
  }
  if (error) {
    return null;
  }

  return (
    <Box
      sx={{
        maxHeight: "100vh",
        overflowY: "scroll",
      }}
    >
      <Stepper orientation="vertical" sx={{ mt: 1 }}>
        {data?.map((item: DocumentHistoryItem) => (
          <ShowDocumentsEditHistoryItem item={item} />
        ))}
      </Stepper>
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
  const timeString = prettifyRelativeTime(item.created_at);

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
    case "favorite":
      return <DocumentHistoryFavorite item={item} pretty_time={timeString} />;
    case "unfavorite":
      return <DocumentHistoryUnfavorite item={item} pretty_time={timeString} />;
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

const iconColor = "secondary";

const ItemText = ({ text }: { text: string }) => {
  return (
    <Typography variant={"body2"} fontSize={"0.8rem"} fontStyle={"italic"}>
      {text}
    </Typography>
  );
};

const ItemLabel = (props: HistoryProps & { action?: string }) => {
  const { item, pretty_time } = props;
  // @ts-ignore
  const fullTime = new Date(Date.parse(item.created_at)).toLocaleString();

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "row",
        justifyContent: "space-between",
      }}
    >
      <Typography variant={"body2"} fontSize={"0.9rem"} component={"span"}>
        {props.action}
      </Typography>
      <Box
        sx={{
          display: "flex",
          flexDirection: "row",
          alignItems: "flex-end",
          gap: "5px",
        }}
      >
        <Tooltip title={`Time: ${fullTime}`}>
          <Typography
            variant="body2"
            fontSize={"0.8rem"}
            fontWeight={"600"}
            gutterBottom
          >
            {pretty_time}
          </Typography>
        </Tooltip>
        <Typography
          variant="body2"
          fontSize={"0.8rem"}
          gutterBottom
          minWidth={"50px"}
        >
          ({item.user})
        </Typography>
      </Box>
    </Box>
  );
};

const DocumentHistoryCreate = (props: HistoryProps) => {
  const { item } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<AddCircleIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Created document"} />
      </StepLabel>
      <StepContent>
        <ItemText text={`Name: ${item.new_value}`} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDescription = (props: HistoryProps) => {
  const { item } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Changed description"} />
      </StepLabel>
      <StepContent>
        <Box
          sx={{
            maxHeight: "40vh",
            overflowY: "scroll",
          }}
        >
          <RecordContextProvider value={props.item}>
            <MarkdownField source={"new_value"} />
          </RecordContextProvider>
        </Box>
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRename = (props: HistoryProps) => {
  const { item } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Renamed document"} />
      </StepLabel>
      <StepContent>
        <ItemText text={`To: ${item.new_value}`} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryAddMetadata = (props: HistoryProps) => {
  const { item } = props;
  let keyId = "";
  let valueId = "";
  let parsed = {};
  let jsonMode = true;
  try {
    parsed = JSON.parse(item.new_value);
    // @ts-ignore
    keyId = get(parsed, "key_id");
    // @ts-ignore
    valueId = get(parsed, "value_id");
  } catch (e) {
    // old mode without json and notation is key_name:value_name
    [keyId, valueId] = item.new_value.split(":");
    jsonMode = false;
  }

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<TagIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Added metadata"} />
      </StepLabel>
      <StepContent>
        {jsonMode ? (
          <Tooltip
            title={`Metadata info: key_id: ${keyId}, value_id: ${valueId}`}
          >
            <ItemText text={"no data"} />
          </Tooltip>
        ) : (
          <ItemText text={"no data"} />
        )}
      </StepContent>
    </Step>
  );
};

const DocumentHistoryRemoveMetadata = (props: HistoryProps) => {
  const { item } = props;
  let keyId = "";
  let valueId = "";
  let parsed = {};
  let jsonMode = true;
  try {
    parsed = JSON.parse(item.old_value);
    // @ts-ignore
    keyId = get(parsed, "key_id");
    // @ts-ignore
    valueId = get(parsed, "value_id");
  } catch (e) {
    [keyId, valueId] = item.old_value.split(":");
    jsonMode = false;
  }
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ContentCut color={iconColor} />}>
        <ItemLabel {...props} action={"Removed metadata"} />
      </StepLabel>
      <StepContent>
        {jsonMode ? (
          <Tooltip
            title={`Metadata info: key_id: ${keyId}, value_id: ${valueId}`}
          >
            <ItemText text={"no data"} />
          </Tooltip>
        ) : (
          <ItemText text={`Metadata: ${keyId}:${valueId}`} />
        )}
      </StepContent>
    </Step>
  );
};

const DocumentHistoryContent = (props: HistoryProps) => {
  const { item } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ArticleIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Changed content"} />
      </StepLabel>
      <StepContent>
        <ItemText text={"Content"} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDate = (props: HistoryProps) => {
  const { item } = props;
  const newDate = new Date(Number(item.new_value) * 1000).toLocaleDateString();

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<ScheduleIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Changed date"} />
      </StepLabel>
      <StepContent>
        <ItemText text={`To: ${newDate}`} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryModifyLinkedDocuments = (props: HistoryProps) => {
  const { item } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<FormatListBulletedIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Modified linked documents"} />
      </StepLabel>
      <StepContent>
        <ItemText text={item.new_value} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryDelete = (props: HistoryProps) => {
  const { item } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<DeleteIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Deleted document"} />
      </StepLabel>
    </Step>
  );
};

const DocumentHistoryRestore = (props: HistoryProps) => {
  const { item } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<RestoreFromTrashIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Restored document"} />
      </StepLabel>
    </Step>
  );
};

const DocumentHistoryLang = (props: HistoryProps) => {
  const { item } = props;
  const newLang = languages[props.item.new_value as keyof typeof languages];

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<TranslateIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Set language"} />
      </StepLabel>
      <StepContent>
        <ItemText text={`To: ${newLang}`} />
      </StepContent>
    </Step>
  );
};

const DocumentHistoryFavorite = (props: HistoryProps) => {
  const { item } = props;
  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<BookmarkIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Added to favorites"} />
      </StepLabel>
      <StepContent></StepContent>
    </Step>
  );
};

const DocumentHistoryUnfavorite = (props: HistoryProps) => {
  const { item } = props;

  return (
    <Step key={`${item.id}`} expanded active completed>
      <StepLabel icon={<BookmarkBorderIcon color={iconColor} />}>
        <ItemLabel {...props} action={"Removed from favorites"} />
      </StepLabel>
      <StepContent></StepContent>
    </Step>
  );
};
