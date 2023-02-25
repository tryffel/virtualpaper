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
  Grid,
  Typography,
} from "@mui/material";
import React from "react";
import { ExpandMore } from "@mui/icons-material";

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

  return (
    <Accordion>
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
        <Typography>
          Started at:
          {props.record.started_at}
        </Typography>
        <Typography>
          Stopped at:
          {props.record.stopped_at}
        </Typography>
      </AccordionDetails>
    </Accordion>
  );
}

export const DocumentTopRow = () => {
  const record = useRecordContext();

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = () => mimetypeToColor(record?.mimetype);
  const getMimeTypeName = (): string => mimetypeToText(record?.mimetype);

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
          sx={{ m: 1 }}
        />
        <Badge
          label={getMimeTypeName()}
          variant="filled"
          color={getMimetypeColor()}
          style={{ top: "4px", right: "16px" }}
          sx={{ m: 1 }}
        />
      </Grid>
      <Grid item xs={12}>
        <DocumentIdField />
      </Grid>
    </Grid>
  );
};

const DocumentIdField = () => {
  const record = useRecordContext();

  const id = record ? record.id : "";

  return (
    <div style={{ marginLeft: "10px", fontWeight: 100, fontSize: "small" }}>
      <span style={{ userSelect: "none" }}>Id:</span> <span>{id}</span>
    </div>
  );
};
