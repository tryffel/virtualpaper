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
  RaRecord,
  RecordContextProvider,
  useDataProvider,
  useGetOne,
  useRecordContext,
} from "react-admin";

import CancelIcon from "@mui/icons-material/Cancel";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";

import {
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  TextField,
  Grid,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";

import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import SettingsIcon from "@mui/icons-material/Settings";
import NotInterestedIcon from "@mui/icons-material/NotInterested";
import { DocumentCard } from "@components/document/card";

interface ConditionResult {
  condition_id: number;
  condition_type: string;
  matched: boolean;
  skipped: boolean;
}

interface ActionResult {
  action_id: number;
  action_type: string;
  skipped: boolean;
}

interface RuleTestResult {
  conditions: ConditionResult[];
  actions: ActionResult[];
  rule_id: number;
  matched: boolean;
  took_ms: number;
  log: string;
  error: string;
  started_at: Date;
  stopped_at: Date;
  condition_output: string[][];
  action_output: string[][];
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

const TestDialog = (props: {
  open: boolean;
  onClose: () => void;
  record: RaRecord;
}) => {
  const record = useRecordContext();
  const [scroll] = React.useState("paper");

  const dataProvider = useDataProvider();

  const { onClose, open } = props;
  const handleClose = () => {
    onClose();
  };

  const [documentId, setDocumentId] = React.useState("");
  const [result, setResult] = React.useState<RuleTestResult | null>();

  const { data, isSuccess, refetch, isError, failureCount } = useGetOne(
    "documents",
    {
      id: documentId,
      meta: {
        noVisit: true,
      },
    },
  );

  const onDocIdchanged = (e: any) => {
    const raw = e.target.value;
    const id = raw.trim();
    setDocumentId(id);
    refetch();
  };

  const handleClear = () => {
    setResult(null);
    setDocumentId("");
  };

  const exec = () => {
    dataProvider
      .testRule("processing/rules", {
        id: record.id,
        data: { document_id: documentId },
      })
      .then((data: { data: RuleTestResult }) => {
        setResult(data.data);
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
            sx={{ minWidth: "75%" }}
            id="document_id"
            label="Document id"
            variant="outlined"
            onChange={onDocIdchanged}
            color={isError || failureCount > 1 || true ? "error" : "primary"}
          />
          <Button onClick={exec} variant="contained" sx={{ mt: 1, ml: 1 }}>
            <Typography>Run test</Typography>
          </Button>

          {isSuccess && data && documentId && (
            <RecordContextProvider value={data}>
              <DocumentCard record={data} />
            </RecordContextProvider>
          )}

          {result ? (
            <>
              <Grid>
                <StatusRow match={result.matched} />
              </Grid>
              <ConditionList
                conditions={result.conditions}
                logs={result.condition_output}
              />
              {result.matched && (
                <ActionList
                  actions={result.actions}
                  logs={result.action_output}
                />
              )}
              <Log log={result && result.log} />
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

export default TestButton;

const StatusRow = (props: { match: boolean }) => {
  const { match } = props;

  return (
    <Grid container flexDirection="row" sx={{ pt: 1 }}>
      <Typography variant="body1" color="textPrimary" sx={{ mt: 0.2, mr: 1 }}>
        Result:
      </Typography>
      {match ? (
        <>
          <Typography variant="body2" sx={{ mr: 1, mt: 0.4 }}>
            Document matched
          </Typography>
          <CheckCircleIcon color="success" />
        </>
      ) : (
        <>
          <Typography variant="body2" sx={{ mr: 1, mt: 0.4 }}>
            No match
          </Typography>
          <CancelIcon color="error" />
        </>
      )}
    </Grid>
  );
};

const Log = (props: { log: string }) => {
  const { log } = props;
  const rows = log.split("\n");

  return (
    <Accordion>
      <AccordionSummary
        expandIcon={<ExpandMoreIcon />}
        style={{ marginTop: "1rem" }}
      >
        <Typography variant="h5" color="textPrimary" sx={{ mt: 1 }}>
          Processing log:
        </Typography>
      </AccordionSummary>
      <AccordionDetails></AccordionDetails>
      <ol>
        {rows.map((line) =>
          line !== "" ? (
            <li>
              <Typography variant="body2">{line}</Typography>
            </li>
          ) : null,
        )}
      </ol>
    </Accordion>
  );
};

const ConditionList = (props: {
  conditions: ConditionResult[];
  logs: string[][];
}) => {
  const { conditions, logs } = props;

  return (
    <Grid paddingTop={"1rem"}>
      <Typography variant="h5" color="textPrimary">
        Conditions
      </Typography>

      {conditions.map((condition, index) => (
        <Condition index={index + 1} condition={condition} logs={logs[index]} />
      ))}
    </Grid>
  );
};

const Condition = (props: {
  index: number;
  condition: ConditionResult;
  logs: string[];
}) => {
  const { index, condition, logs } = props;

  return (
    <Grid container flexDirection={"column"} paddingTop={"1rem"}>
      <Grid
        container
        flexDirection={"row"}
        justifyContent={"flex-start"}
        position={"relative"}
      >
        <Grid item sx={{ mr: "5px" }}>
          {index}.
        </Grid>
        <Grid item>
          <Typography variant="body1" sx={{ fontWeight: 600 }}>
            {condition.condition_type}
          </Typography>
        </Grid>
        <Grid
          item
          justifyContent={"flex-end"}
          position={"absolute"}
          right={"0px"}
        >
          {condition.skipped ? (
            <NotInterestedIcon color={"disabled"} />
          ) : condition.matched ? (
            <CheckCircleIcon color="success" />
          ) : (
            <CancelIcon color="error" />
          )}
        </Grid>
      </Grid>
      {logs && (
        <Grid item>
          <ol style={{ margin: "0 auto" }}>
            {logs.map((entry) => (
              <li>{entry}</li>
            ))}
          </ol>
        </Grid>
      )}
    </Grid>
  );
};

const ActionList = (props: { actions: ActionResult[]; logs: string[][] }) => {
  const { actions, logs } = props;

  return (
    <Grid paddingTop={"1rem"}>
      <Typography variant="h5" color="textPrimary">
        Actions
      </Typography>

      {actions.map((action, index) => (
        <Action index={index + 1} action={action} logs={logs[index]} />
      ))}
    </Grid>
  );
};

const Action = (props: {
  index: number;
  action: ActionResult;
  logs: string[];
}) => {
  const { index, action, logs } = props;

  return (
    <Grid container flexDirection={"column"} paddingTop={"1rem"}>
      <Grid
        container
        flexDirection={"row"}
        justifyContent={"flex-start"}
        position={"relative"}
      >
        <Grid item sx={{ mr: "5px" }}>
          {index}.
        </Grid>
        <Grid item>
          <Typography variant="body1" sx={{ fontWeight: 600 }}>
            {action.action_type}
          </Typography>
        </Grid>
      </Grid>
      {logs && (
        <Grid item>
          <ol style={{ margin: "0 auto" }}>
            {logs.map((entry) => (
              <li>{entry}</li>
            ))}
          </ol>
        </Grid>
      )}
    </Grid>
  );
};
