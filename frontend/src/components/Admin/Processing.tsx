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
import { Confirm, useDataProvider, useNotify } from "react-admin";

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
} from "@mui/material";
import { fstat } from "fs";

const Processing = () => {
  return (
    <Box>
      <RequestSingleDocumentProcessing />
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
  const [stepOk, setStepOk] = useState(true);
  const [userId, setUserId] = useState("");

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

  const notify = useNotify();

  const onChangeStep = (e: any) => {
    setStep(e.target.value);
    let ok = false;
    steps.map((step) => {
      // @ts-ignore
      if (step.id == e.target.value) {
        ok = true;
      }
    });
    setStepOk(ok);
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
        notify("Processing scheduled");
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
      />

      <TextField
        id="user"
        label="UserId"
        variant="outlined"
        value={userId}
        onChange={onChangeUserId}
      />
      <TextField
        id="step"
        label="Step"
        variant="outlined"
        value={step}
        onChange={onChangeStep}
        color={stepOk ? "primary" : "secondary"}
      />

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

export default Processing;
