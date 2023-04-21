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
  BooleanField,
  Button, CreateButton,
  Datagrid,
  EditButton,
  List,
  TopToolbar,
  useRecordContext,
} from "react-admin";
import ReorderIcon from "@mui/icons-material/Reorder";

import { Box, Chip, Grid, Typography } from "@mui/material";
import { MarkdownField } from "../Markdown";
import get from "lodash/get";
import { EmptyResourcePage } from "../primitives/EmptyPage";
import { ReorderRulesDialog, RuleTitle } from "./Reorder";

export const RuleList = () => (
  <List empty={<EmptyRuleList />} actions={<RuleListActions />}>
    <Datagrid bulkActionButtons={false} expand={ExpandRule}>
      <RuleTitle />
      <BooleanField label="Enabled" source="enabled" />
      <EditButton />
    </Datagrid>
  </List>
);

const RuleModeField = (props: any) => {
  const { source } = props;
  const record = useRecordContext(props);
  const value = get(record, source);

  return <Chip label={value === "match_all" ? "Match all" : "Match any"} />;
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
            <Typography variant="body2">Description</Typography>
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
      <CreateButton/>
    </TopToolbar>
  );
};
