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
  CardContentInner,
  SelectInput,
} from "react-admin";

import {
  Typography,
  TextField,
  Box,
  Button,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
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
import { fstat } from "fs";
import { RequestIndexSelect } from "../Documents/RequestIndexing";

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
    id: "rules",
    name: "UserRules",
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
      // @ts-ignore
      .then((data: { data: any }) => {
        notify(`Processing scheduled`, { type: "success" });
      });
  };

  return (
    <Box>
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
      <Typography variant="h6">Instructions</Typography>
      <Typography variant="body1">
        <ul>
          <li>Force processing single or multiple documents</li>
          <li>
            Please fill either document id or user id. If user id (numeric) is
            passed, all the documetns for the user are processed starting from
            the given step.
          </li>
          <li>
            Allowed steps are:
            <TableContainer component={Paper} elevation={2}>
              <Table sx={{ minWidth: 300 }}>
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
                      sx={{ "&:last-child td, &:last-child th": { border: 0 } }}
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
          <li>
            All subsequent steps are always executed. E.g. if step is set to
            content, rules and then fts are executed as well.
          </li>
        </ul>
      </Typography>

      <Typography>Allowed steps: thumbnail, content, rules or fts</Typography>
      <TextField
        id="document_id"
        label="Document id"
        variant="outlined"
        value={documentId}
        onChange={onChangeDocumentId}
        disabled={!!userId}
      />
      <Typography>OR</Typography>

      <TextField
        id="user"
        label="UserId"
        variant="outlined"
        value={userId}
        onChange={onChangeUserId}
        disabled={!!documentId}
      />
      <RequestIndexSelect step={step} setStep={onChangeStep} enabledAll />
      <Button
        onClick={onClickExec}
        variant="contained"
        disabled={!documentId ? userId === "" : userId !== ""}
      >
        <Typography>Execute</Typography>
      </Button>
    </Box>
  );
};

const ShowDocumentsPendingProcessing = () => {
  const { data, total, isLoading, error, refetch } = useGetList(
    "admin/documents/processing",
    {}
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

export const Runners = (props: any) => {
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  if (!props.record) {
    return <Loading />;
  }

  return (
    <Grid container spacing={2} direction={isSmall ? "column" : "row"}>
      <Grid item xs={12} md={12}>
        <Typography variant="body2">Processing runners</Typography>
      </Grid>

      {props.record.processing_queue.map(
        // @ts-ignore
        (runner) => (
          <Grid item xs={3}>
            <RunnerStatus runner={runner} />
          </Grid>
        )
      )}
    </Grid>
  );
};

const RunnerStatus = (props: any) => {
  if (!props.runner) {
    return null;
  }

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
        title={"Task " + props.runner.task_id + ": " + runnerStatus}
        sx={{
          background: backgroundColor,
        }}
      />
      <CardContent>
        <Typography>
          Queue: ({props.runner.queued} / {props.runner.queue_capacity} )
        </Typography>
        <Typography>
          Document id: {props.runner.processing_document_id}
        </Typography>
        <Typography>
          Step duration: {Math.floor(props.runner.duration_ms / 1000)} s
        </Typography>
      </CardContent>
    </Card>
  );
};

export const SearchEngineStatus = (props: any) => {
  if (!props.status) {
    return <p>Loading</p>;
  }

  let runnerStatus = "";
  let backgroundColor = "";
  if (!props.status.engine_ok) {
    runnerStatus = "Unavailable";
    backgroundColor = "red";
  } else {
    runnerStatus = "running";
    backgroundColor = "";
  }

  return (
    <Card
      key={"search engine status"}
      elevation={3}
      sx={{ margin: "0.5em", display: "inline-block" }}
    >
      <CardHeader
        title={"Search engine status"}
        sx={{
          background: backgroundColor,
        }}
      />
      <CardContent>
        <Typography>
          {props.status.name} {props.status.version}
        </Typography>
        <Typography>Status: {props.status.status}</Typography>
      </CardContent>
    </Card>
  );
};
