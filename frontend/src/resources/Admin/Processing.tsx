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

import React, { useEffect, useState } from "react";
import {
  Confirm,
  useDataProvider,
  useNotify,
  useGetList,
  Loading,
  HttpError,
  useRecordContext,
} from "react-admin";

import {
  Typography,
  TextField,
  Box,
  Button,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Paper,
  Grid,
  Card,
  CardContent,
  CardHeader,
  useMediaQuery,
} from "@mui/material";
import { RequestIndexSelect } from "@components/document/edit/RequestIndexing.tsx";
import OfflinePinIcon from "@mui/icons-material/OfflinePin";
import ReportProblemIcon from "@mui/icons-material/ReportProblem";
import AutorenewIcon from "@mui/icons-material/Autorenew";
import get from "lodash/get";

export const Processing = () => {
  return (
    <Box>
      <RequestSingleDocumentProcessing />
    </Box>
  );
};

export const DocumentList = () => {
  return (
    <Box>
      <ShowDocumentsPendingProcessing />
    </Box>
  );
};

const steps = [
  { id: "thumbnail", name: "Thumbnail", description: "Generate new thumbnail" },
  {
    id: "content",
    name: "Extract content",
    description: "Extract content from raw document",
  },
  {
    id: "detect-language",
    name: "Detect language",
    description: "Detect the language of the document",
  },
  {
    id: "rules",
    name: "User Rules",
    description: "Execute user defined rules and metadata matching",
  },
  {
    id: "fts",
    name: "Index",
    description: "Reindex document in full-text-search engine",
  },
];
//}

const RequestSingleDocumentProcessing = () => {
  const dataProvider = useDataProvider();
  const [documentId, setDocumentId] = useState("");
  const [step, setStep] = useState("content");
  const [userId, setUserId] = useState("");
  const notify = useNotify();

  const [confirmOpen, setConfirmOpen] = useState(false);

  const onClickExec = () => {
    setConfirmOpen(true);
  };

  const onCancel = () => {
    setConfirmOpen(false);
  };

  const onConfirm = () => {
    setConfirmOpen(false);
    exec();
  };

  const onChangeStep = (val: string) => {
    setStep(val);
  };
  const onChangeDocumentId = (e: any) => {
    setDocumentId(e.target.value);
  };

  const onChangeUserId = (e: any) => {
    if (!isNaN(e.target.value)) {
      setUserId(e.target.value);
    }
  };

  const exec = () => {
    dataProvider
      .adminRequestProcessing({
        data: {
          user_id: userId !== "" ? parseInt(userId) : 0,
          document_id: documentId !== "" ? documentId : "",
          from_step: step !== "" ? step : "",
        },
      })
      .then((data: any) => {
        console.log("received data", data);
        return data;
      })
      // @ts-ignore
      .then((data: { data: any }) => {
        notify(`Processing scheduled`, { type: "success" });
      })
      .catch((err: HttpError) => {
        notify(err.message, { type: "error" });
      });
  };

  return (
    <Grid container spacing={2}>
      <Confirm
        isOpen={confirmOpen}
        title={
          userId
            ? "Request processing all documents for a user"
            : "Request processing for a single document"
        }
        content={
          userId
            ? "Are you sure you want to request processing?" +
              " This operation may take a while if the user has great number of documents."
            : "Are you sure you want to request processing?"
        }
        onConfirm={onConfirm}
        onClose={onCancel}
      />
      <Grid item xs={12}>
        <Typography variant="h6">Instructions</Typography>
        <Typography variant="body1">
          <ul>
            <li>
              Please fill either document id or user id. If user id (numeric) is
              passed, all the documents for the user are processed starting from
              the given step.
            </li>
            <li>
              Allowed steps are:
              <TableContainer component={Paper} elevation={2}>
                <Table sx={{ minWidth: 300 }} size={"small"}>
                  <TableHead>
                    <TableRow>
                      <TableCell>#</TableCell>
                      <TableCell align="left">Name</TableCell>
                      <TableCell align="left">Description</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {steps.map((step, index) => (
                      <TableRow
                        key={index}
                        sx={{
                          "&:last-child td, &:last-child th": { border: 0 },
                        }}
                      >
                        <TableCell>{index + 1}</TableCell>
                        <TableCell align="left" component="th" scope="row">
                          {step.name}
                        </TableCell>
                        <TableCell align="left">{step.description}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </li>
          </ul>
        </Typography>
      </Grid>
      <Grid item xs={12} sm={8} md={8} lg={4}>
        <TextField
          id="document_id"
          label="Document id"
          variant="outlined"
          value={documentId}
          onChange={onChangeDocumentId}
          disabled={!!userId}
          fullWidth
        />
      </Grid>
      <Grid
        item
        xs={1}
        sm={1}
        md={1}
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "start",
        }}
      >
        <Typography pl={"10px"}>OR</Typography>
      </Grid>
      <Grid item xs={12} sm={8} md={8} lg={4}>
        <TextField
          id="user"
          label="UserId"
          variant="outlined"
          value={userId}
          onChange={onChangeUserId}
          disabled={!!documentId}
          fullWidth
        />
      </Grid>
      <Grid
        item
        xs={12}
        sm={6}
        md={6}
        lg={4}
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "start",
        }}
      >
        <RequestIndexSelect step={step} setStep={onChangeStep} enabledAll />
      </Grid>
      <Grid item xs={12}>
        <Button
          onClick={onClickExec}
          variant="contained"
          disabled={!documentId ? userId === "" : userId !== ""}
        >
          <Typography>Execute</Typography>
        </Button>
      </Grid>
    </Grid>
  );
};

const ShowDocumentsPendingProcessing = () => {
  const { data, total, isLoading, error, refetch } = useGetList(
    "admin/documents/processing",
    {},
  );

  let interval: number = 0;

  useEffect(() => {
    // @ts-ignore
    interval = setInterval(() => {
      if (!isLoading) {
        refetch();
      }
    }, 5000);

    return function cleanup() {
      clearInterval(interval);
    };
  });

  if (isLoading) {
    return <Loading />;
  }
  if (error) {
    return <p>ERROR</p>;
  }

  return (
    <>
      <Typography variant="body1">
        Total steps waiting for processing: {total}
      </Typography>
      <Typography variant="body2">Processing queue</Typography>
      <TableContainer component={Paper} elevation={2}>
        <Table sx={{ minWidth: 300 }}>
          <TableHead>
            <TableRow>
              <TableCell>#</TableCell>
              <TableCell align="left">Document id</TableCell>
              <TableCell align="left">Step</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data
              ? // @ts-ignore
                data.map((step, index) => (
                  <TableRow
                    key={index}
                    sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
                  >
                    <TableCell>{index + 1}</TableCell>
                    <TableCell align="left" component="th" scope="row">
                      {step.id}
                    </TableCell>
                    <TableCell align="left">{step.step}</TableCell>
                  </TableRow>
                ))
              : null}
          </TableBody>
        </Table>
      </TableContainer>
    </>
  );
};

export const Runners = () => {
  const record = useRecordContext();
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  if (!record) {
    return null;
  }

  return (
    <Grid container spacing={2} direction={isSmall ? "column" : "row"}>
      <Grid item xs={12}>
        <Typography variant={"h5"}>
          Processing tasks ({get(record, "processing_queue").length})
        </Typography>
      </Grid>
      {record.processing_queue.map(
        // @ts-ignore
        (runner) => (
          <Grid item xs={3} key={runner.id}>
            <RunnerStatus runner={runner} />
          </Grid>
        ),
      )}
    </Grid>
  );
};

type RunnerStatus = "running" | "idle" | "not-running";

const RunnerIcons: Record<RunnerStatus, React.ReactNode> = {
  running: <AutorenewIcon color={"warning"} />,
  idle: <OfflinePinIcon color={"info"} />,
  "not-running": <ReportProblemIcon color={"error"} />,
};

const RunnerStatus = (props: any) => {
  if (!props.runner) {
    return null;
  }

  const getStatus = (): RunnerStatus => {
    if (!props.runner.task_running) {
      return (runnerStatus = "not-running");
    }
    if (props.runner.processing_ongoing) {
      return "running";
    }
    return "idle";
  };

  let runnerStatus = "";
  let backgroundColor = "";
  if (!props.runner.task_running) {
    runnerStatus = "not running";
    backgroundColor = "red";
  } else if (props.runner.processing_ongoing) {
    runnerStatus = "running";
    backgroundColor = "orange";
  } else {
    runnerStatus = "idle";
  }

  return (
    <Card key={props.runner.task_id} elevation={3}>
      <CardHeader
        avatar={RunnerIcons[getStatus()]}
        title={
          <Typography variant={"h6"} style={{ display: "inline" }}>
            {parseInt(props.runner.task_id) + 1}: {runnerStatus}
          </Typography>
        }
        sx={{
          background: backgroundColor,
        }}
      />
      <CardContent>
        <Typography variant={"body2"}>
          Queue: ({props.runner.queued} / {props.runner.queue_capacity} )
        </Typography>
        <Typography variant={"body2"}>
          Document: {props.runner.processing_document_id}
        </Typography>
        <Typography variant={"body2"}>
          Step duration: {Math.floor(props.runner.duration_ms / 1000)} s
        </Typography>
      </CardContent>
    </Card>
  );
};

export const SearchEngineStatus = () => {
  const record = useRecordContext();

  const ok = get(record, "search_engine_status.engine_ok");
  const status = get(record, "search_engine_status.status");
  const version = get(record, "search_engine_status.version");

  const color = ok && status === "available" ? "info" : "error";
  const statusLabel = ok && status === "available" ? "ok" : status;

  return (
    <Grid container spacing={1}>
      <Grid item xs={12}>
        <Typography variant={"h5"}>Search engine</Typography>
      </Grid>
      <Grid item xs={12}>
        <Card
          key={"search engine"}
          elevation={3}
          sx={{ display: "inline-block" }}
        >
          <CardHeader
            avatar={
              ok ? (
                <OfflinePinIcon color={"info"} />
              ) : (
                <ReportProblemIcon color={"error"} />
              )
            }
            title={<Typography variant={"h6"}>Meilisearch</Typography>}
            sx={{
              background: color,
            }}
          />
          <CardContent>
            <Typography>Version: {version}</Typography>
            <Typography>Status: {statusLabel}</Typography>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
};
