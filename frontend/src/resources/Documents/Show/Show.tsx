import {
  Loading,
  TextField,
  useGetManyReference,
  useRecordContext,
} from "react-admin";
import { Badge, mimetypeToColor, mimetypeToText } from "../DocumentCard";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Grid,
  Tooltip,
  Typography,
} from "@mui/material";
import React from "react";
import { ExpandMore } from "@mui/icons-material";
import get from "lodash/get";
import { languages } from "../../../languages";
import {
  PrettifyAbsoluteTime,
  PrettifyTimeInterval,
} from "../../../components/util";

function getLanguageLabel(code: string): string {
  const lang = languages[code as keyof typeof languages];
  return lang ? (lang as string) : code;
}

export function DocumentJobsHistory() {
  const record = useRecordContext();
  const { data, isLoading, error } = useGetManyReference("document/jobs", {
    target: "id",
    id: record?.id,
    sort: {
      field: "timestamp",
      order: "ASC",
    },
  });

  if (isLoading) {
    return <Loading />;
  }
  if (error) {
    return null;
    // return <Error />;
  }

  if (data !== undefined) {
    return (
      <div>
        {data.map((index) => (
          <DocumentJobListItem record={index} />
        ))}
      </div>
    );
  }
  return null;
}

function DocumentJobListItem(props: any) {
  if (!props.record) {
    return null;
  }
  const ok = props.record.status === "Finished";
  let style = {};
  let prefix = "";
  if (props.record.status === "Finished") {
  } else if (props.record.status === "Running") {
    style = { fontStyle: "italic", background: "#ff0" };
    prefix = "Running";
  } else if (props.record.status === "Error") {
    style = { fontStyle: "italic", background: "red" };
    prefix = "Error";
  }
  const startTime = PrettifyAbsoluteTime(props.record.started_at);
  const took = PrettifyTimeInterval(
    props.record.started_at,
    props.record.stopped_at
  );

  return (
    <Accordion sx={{ minWidth: "15em" }}>
      <AccordionSummary expandIcon={<ExpandMore />} style={style}>
        <Typography>
          {prefix} {props.record.message}
        </Typography>
      </AccordionSummary>
      <AccordionDetails style={{ flexDirection: "column" }}>
        <Typography>
          Status:
          {props.record.status}
        </Typography>
        <Typography>
          Job id:
          {props.record.id}
        </Typography>

        <Tooltip title={`Time: ${took !== "" ? took : "< 1 sec"}`}>
          <Typography>Started at: {startTime}</Typography>
        </Tooltip>
      </AccordionDetails>
    </Accordion>
  );
}

export const DocumentTopRow = () => {
  const record = useRecordContext();
  const isDeleted = get(record, "deleted_at") !== null;
  const hasLang = get(record, "lang") !== "";

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = () => mimetypeToColor(record?.mimetype);
  const getMimeTypeName = (): string => mimetypeToText(record?.mimetype);
  const getLang = () => getLanguageLabel(record?.lang);

  return (
    <Grid container justifyContent={"align-center"}>
      <Grid item>
        <TextField source="name" label="" style={{ fontSize: "2em" }} />
      </Grid>

      <Grid item xs={12}>
        <Badge
          label={getDateString()}
          variant="outlined"
          color={"primary"}
          style={{
            top: "4px",
            left: "16px",
            background: "white",
          }}
          sx={{ m: 0.5 }}
        />
        {hasLang && (
          <Badge
            label={getLang()}
            variant="outlined"
            color={"primary"}
            sx={{ m: 0.5 }}
          />
        )}
        <Badge
          label={getMimeTypeName()}
          variant="filled"
          color={getMimetypeColor()}
          style={{ top: "4px", right: "16px" }}
          sx={{ m: 1 }}
        />
        {isDeleted && (
          <Badge
            label={"Document is deleted"}
            variant="filled"
            color={"error"}
            style={{ top: "4px", right: "16px" }}
            sx={{ m: 1 }}
          />
        )}
      </Grid>
      <Grid item xs={12}>
        <DocumentIdField />
      </Grid>
    </Grid>
  );
};

export const DocumentBasicInfo = () => {
  const record = useRecordContext();
  const isDeleted = get(record, "deleted_at") !== null;
  const hasLang = get(record, "lang") !== "";

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = () => mimetypeToColor(record?.mimetype);
  const getMimeTypeName = (): string => mimetypeToText(record?.mimetype);
  const getLang = () => getLanguageLabel(record?.lang);

  return (
    <Box>
      <Badge
        label={getDateString()}
        variant="outlined"
        color={"primary"}
        style={{
          top: "4px",
          left: "16px",
          background: "white",
          marginLeft: 0,
        }}
        sx={{ m: 0.5 }}
      />
      {hasLang && (
        <Badge
          label={getLang()}
          variant="outlined"
          color={"primary"}
          sx={{ m: 0.5 }}
        />
      )}
      <Badge
        label={getMimeTypeName()}
        variant="filled"
        color={getMimetypeColor()}
        style={{ top: "4px", right: "16px" }}
        sx={{ m: 1 }}
      />
      {isDeleted && (
        <Badge
          label={"Document is deleted"}
          variant="filled"
          color={"error"}
          style={{ top: "4px", right: "16px" }}
          sx={{ m: 1 }}
        />
      )}
    </Box>
  );
};

export const DocumentTitle = () => {
  return (
    <TextField
      source="name"
      label=""
      style={{ fontSize: "2em", paddingBottom: "0.8em" }}
    />
  );
};

export const DocumentIdField = () => {
  const record = useRecordContext();

  const id = record ? record.id : "";

  return (
    <div style={{ marginLeft: "10px", fontWeight: 100, fontSize: "small" }}>
      <span style={{ userSelect: "none" }}>Id: </span>
      <span>{id}</span>
    </div>
  );
};
