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
  ArrayInput,
  BooleanInput,
  DateField,
  Edit,
  RadioButtonGroupInput,
  SimpleForm,
  TextInput,
  useRecordContext,
} from "react-admin";

import { Box, Grid, Typography, useTheme } from "@mui/material";
import { MarkdownInput } from "../../Markdown";
import TestButton from "../Test";
import { RuleEditHelp } from "./Help";
import { ConditionEdit } from "./Condition";
import { ActionEdit } from "./Action";

export const RuleEdit = () => {
  const theme = useTheme();
  const record = useRecordContext();

  return (
    <Edit title={"Edit process rule"}>
      <SimpleForm>
        <div>
          <Grid container spacing={2}>
            <Grid item xs={12} md={8} lg={12}>
              <Box display={{ xs: "block", sm: "flex" }}>
                <Typography variant="h5">Edit Processing Rule</Typography>
                <RuleEditHelp />
                <TestButton record={record} />
              </Box>
              <Box display={{ xs: "block", sm: "flex" }}>
                <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
                  <TextInput source="name" />
                </Box>
                <Box flex={1} mr={{ xs: 0, sm: "0.5em" }}>
                  <BooleanInput label="Enabled" source="enabled" />
                </Box>
              </Box>
              <Typography variant="body2">Created at</Typography>
              <DateField source="created_at" />

              <Box display={{ xs: "block", sm: "flex" }}>
                <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                  <MarkdownInput source="description" />
                  <MatchTypeSelectInput source="mode" />
                  <Typography variant="h5">Rule Conditions</Typography>
                  <ArrayInput
                    source="conditions"
                    label=""
                    sx={{
                      background: theme.palette.background.default,
                      border: "1px",
                      borderRadius: "5px",
                      margin: "1em",
                      padding: "1em",
                      boxShadow: "1",
                    }}
                  >
                    <ConditionEdit />
                  </ArrayInput>
                  <Typography variant="h5">Rule Actions</Typography>
                  <ArrayInput
                    source="actions"
                    label=""
                    sx={{
                      background: theme.palette.background.default,
                      border: "1px",
                      borderRadius: "5px",
                      margin: "1em",
                      padding: "1em",
                      boxShadow: "1",
                    }}
                  >
                    <ActionEdit />
                  </ArrayInput>
                </Box>
              </Box>
            </Grid>
          </Grid>
        </div>
      </SimpleForm>
    </Edit>
  );
};

interface SourceProps {
  source: string;
}

const MatchTypeSelectInput = (props: SourceProps) => {
  const { source } = props;
  return (
    <RadioButtonGroupInput
      label="Match conditions"
      source={source}
      fullWidth={true}
      choices={[
        { id: "match_all", name: "Match all" },
        { id: "match_any", name: "Match any" },
      ]}
    />
  );
};

// case-insensitive
// inverted
// regex
// on-condition
