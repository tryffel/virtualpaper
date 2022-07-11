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
import { useDataProvider, useNotify } from "react-admin";

import {
  Typography,
  TextField,
  Box,
  Button,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
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
  { id: "thumbnail", name: "Thumbnail" },
  { id: "content", name: "Extract content" },
  { id: "rules", name: "UserRules" },
  { id: "fts", name: "Index" },
];
//}

const RequestSingleDocumentProcessing = () => {
  const dataProvider = useDataProvider();
  const [documentId, setDocumentId] = useState("");
  const [step, setStep] = useState("content");
  const [stepOk, setStepOk] = useState(true);
  const [userId, setUserId] = useState("");

  const notify = useNotify();

  const onChangeStep = (e: any) => {
    setStep(e.target.value);
    let ok = false;
    steps.map((step) => {
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
      <Typography variant="h6">
        Force processing of one or more documents from given step
      </Typography>
      <Typography variant="body1">
        Please fill either document id or user id. If user id is set, all
        documents of that user are processed starting from the given step.
      </Typography>

      <Typography>Allowed steps: thumbnail|content|rules|fts</Typography>
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

      <Button onClick={exec} variant="contained">
        <Typography>Execute</Typography>
      </Button>
    </Box>
  );
};

export default Processing;
