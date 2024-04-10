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
  CreateButton,
  Datagrid,
  EditButton,
  List,
  SearchInput,
  SelectInput,
  TopToolbar,
  useRecordContext,
} from "react-admin";
import ReorderIcon from "@mui/icons-material/Reorder";

import { Box, Chip, Grid, Tooltip, Typography } from "@mui/material";
import { MarkdownField } from "@components/markdown";
import get from "lodash/get";
import { EmptyResourcePage } from "@components/primitives/EmptyPage.tsx";
import { ReorderRulesDialog, RuleTitle } from "./Reorder";
import ControlPointIcon from "@mui/icons-material/ControlPoint";
import BorderColorIcon from "@mui/icons-material/BorderColor";

export const RuleList = () => (
  <List
    empty={<EmptyRuleList />}
    actions={<RuleListActions />}
    filters={[
      <SearchInput source={"q"} alwaysOn />,
      <SelectInput
        label={"Enabled"}
        source={"enabled"}
        alwaysOn
        choices={[
          { id: "true", name: "Enabled" },
          {
            id: "false",
            name: "Disabled",
          },
        ]}
      />,
    ]}
  >
    <Datagrid bulkActionButtons={false} expand={ExpandRule}>
      <RuleTitle source={"Name"} />
      <RuleTriggerField source={"triggers"} />
      <EditButton />
    </Datagrid>
  </List>
);

const RuleModeField = (props: { source: string }) => {
  const { source } = props;
  const record = useRecordContext(props);
  const value = get(record, source);

  return <Chip label={value === "match_all" ? "Match all" : "Match any"} />;
};

const RuleTriggerField = (_: { source: string }) => {
  const record = useRecordContext();

  const hasCreate =
    get(record, "triggers")?.filter(
      (entry: string) => entry === "document-create",
    ).length > 0;
  const hasUpdate =
    get(record, "triggers")?.filter(
      (entry: string) => entry === "document-update",
    ).length > 0;

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "row",
        gap: "10px",
      }}
    >
      {hasCreate ? (
        <Tooltip title={"Run after new document has been created"}>
          <ControlPointIcon
            color={"secondary"}
            sx={{ height: "20px", width: "20px" }}
          />
        </Tooltip>
      ) : (
        <div style={{ width: "20px" }}></div>
      )}
      {hasUpdate ? (
        <Tooltip title={"Run after existing document has been updated"}>
          <BorderColorIcon
            color={"secondary"}
            sx={{ height: "20px", width: "20px" }}
          />
        </Tooltip>
      ) : (
        <div style={{ width: "20px" }}></div>
      )}
    </Box>
  );
};

const ChildCounterField = (props: any) => {
  const { source } = props;
  const record = useRecordContext(props);
  const value = get(record, source);

  return record ? (
    <Typography component="span" variant="body2">
      {value ? value.length : ""}
    </Typography>
  ) : null;
};

const ExpandRule = () => {
  return (
    <Grid container>
      <Grid item xs={6} md={6} lg={6}>
        <Box display={{ xs: "block", sm: "flex" }}>
          <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
            <MarkdownField label="Description" source="description" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Mode</Typography>
            <RuleModeField source="mode" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Conditions</Typography>
            <ChildCounterField source="conditions" />
          </Box>
          <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
            <Typography variant="body2">Actions</Typography>
            <ChildCounterField source="actions" />
          </Box>
        </Box>
      </Grid>
    </Grid>
  );
};

const EmptyRuleList = () => {
  return (
    <EmptyResourcePage
      title={"No processing rules"}
      subTitle={"Do you want to add one?"}
    />
  );
};

const RuleListActions = () => {
  const [modalOpen, setModalOpen] = React.useState(false);

  return (
    <TopToolbar>
      <Button label={"Reorder"} onClick={() => setModalOpen(true)}>
        <ReorderIcon />
      </Button>
      <ReorderRulesDialog setModalOpen={setModalOpen} modalOpen={modalOpen} />
      <CreateButton />
    </TopToolbar>
  );
};
