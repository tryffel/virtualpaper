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

import * as React from "react";

import {
  Button,
  useDataProvider,
  useGetOne,
  useRecordContext,
} from "react-admin";

import {
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  TextField,
} from "@mui/material";

import { Settings as SettingsIcon } from "@mui/icons-material";

interface RuleTestResult {
  conditions: {
    condition_id: number;
    matched: boolean;
  }[];
  rule_id: number;
  matched: boolean;
  took_ms: number;
  log: string;
  error: string;
  started_at: Date;
  stopped_at: Date;
}

const TestButton = (record: any) => {
  const [open, setOpen] = React.useState(false);

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  return (
    <div>
      <Button
        label="Test rule"
        size="small"
        alignIcon="left"
        onClick={handleClickOpen}
      >
        <SettingsIcon />
      </Button>
      <TestDialog open={open} onClose={handleClose} record={record} />
    </div>
  );
};

const TestDialog = (props: any) => {
  const record = useRecordContext();
  const [scroll] = React.useState("paper");

  const dataProvider = useDataProvider();

  const { onClose, open } = props;
  const handleClose = () => {
    onClose();
  };

  const [documentId, setDocumentId] = React.useState("");
  const [result, setResult] = React.useState<RuleTestResult>();

  const [textResult, setTextResult] = React.useState("");

  const { data, refetch, isError, isLoadingError, failureCount } = useGetOne(
    "documents",
    {
      id: documentId,
      meta: {
        noVisit: true,
      },
    }
  );

  console.log("fetch error", isError, isLoadingError, failureCount);

  const onDocIdchanged = (e: any) => {
    const raw = e.target.value;
    const id = raw.trim();
    setDocumentId(id);
    refetch();
  };

  const handleClear = () => {
    // @ts-ignore
    setResult(null);
    setTextResult("");
  };

  const exec = () => {
    // @ts-ignore
    dataProvider
      .testRule("processing/rules", {
        id: record.id,
        data: { document_id: documentId },
      })
      // @ts-ignore
      .then((data: { data: RuleTestResult }) => {
        setResult(data.data);
        const splits = data.data.log.split("\n");
        setTextResult(data.data.log);
        //setTextResult(splits.j)
      });
  };

  return (
    <Dialog
      onClose={handleClose}
      aria-labelledby="simple-dialog-title"
      open={open}
    >
      <DialogTitle id="simple-dialog-title">Test Processing Rule</DialogTitle>
      <DialogContent dividers={scroll === "paper"}>
        <DialogContentText>
          <Typography variant="body2">
            <p>
              Processing rule can be tested against a document to see if the
              document matches the conditions that the rule has been configured
              with.
            </p>
            <p>
              No changes to the document will be made. This tool is only for
              debugging purposes.
            </p>

            <p>
              To test a rule enter a document's id below and click{" "}
              <em>Run test</em>. After running the test a list of entries that
              describes the execution, is shown below.
            </p>
          </Typography>

          <TextField
            helperText={
              (isError || failureCount > 1) && "Id does not match any document"
            }
            fullWidth
            id="document_id"
            label="Document id"
            variant="outlined"
            // @ts-ignore
            onChange={onDocIdchanged}
            color={isError ? "error" : "primary"}
          />

          {
            // @ts-ignore
            data ? (
              <Typography variant="body2" sx={{ pt: 1 }}>
                Document name: '{data.name}'
              </Typography>
            ) : null
          }

          <Button onClick={exec} variant="contained" sx={{ mt: 1 }}>
            <Typography>Run test</Typography>
          </Button>

          {result ? (
            <>
              <Typography variant="h5" color="textPrimary" sx={{ pt: 1 }}>
                Results
              </Typography>
              <p>Match: {result.matched ? "yes" : "no"}</p>

              <Typography variant="h5" color="textPrimary">
                Log
              </Typography>
              <ul>
                {" "}
                {result
                  ? result.log
                      .split("\n")
                      // @ts-ignore
                      .map((line) => (line !== "" ? <li>{line}</li> : null))
                  : null}
              </ul>

              <Typography variant="h5" color="textPrimary">
                Conditions
              </Typography>
              <ul>
                {result.conditions.map((condition, index) => (
                  <li>
                    condition {index + 1} matched:{" "}
                    {condition.matched ? "yes" : "no"}
                  </li>
                ))}
              </ul>
            </>
          ) : null}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button variant="contained" onClick={handleClear}>
          <Typography>Clear</Typography>
        </Button>
        <Button variant="outlined" onClick={handleClose}>
          <Typography>Close</Typography>
        </Button>
      </DialogActions>
    </Dialog>
  );
};

const Test = () => {
  return <p>kakka pissa</p>;
};

export default TestButton;
