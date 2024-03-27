import {
  PrettifyAbsoluteTime,
  PrettifyTimeInterval,
} from "@components/util.ts";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Tooltip,
  Typography,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import { Loading, useGetManyReference, useRecordContext } from "react-admin";

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
  let style = {};
  let prefix = "";
  if (props.record.status === "Finished") {
    // pass
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
    props.record.stopped_at,
  );

  return (
    <Accordion sx={{ minWidth: "15em" }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon />} style={style}>
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
