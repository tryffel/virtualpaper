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

import {
  ArrayInput,
  BooleanInput,
  Edit,
  Form,
  Labeled,
  RadioButtonGroupInput,
  SaveButton,
  SelectArrayInput,
  TextInput,
  Toolbar,
  useRecordContext,
} from "react-admin";

import { Box, Grid, Typography, useTheme } from "@mui/material";
import TestButton from "../Test";
import { RuleEditHelp } from "./Help";
import { ConditionEdit } from "./Condition";
import { ActionEdit } from "./Action";
import { MarkdownInput } from "../../../components/markdown";
import { TimestampField } from "@components/primitives/TimestampField.tsx";

export const RuleEdit = () => {
  const theme = useTheme();
  const record = useRecordContext();

  return (
    <Edit title={"Edit process rule"}>
      <Form>
        <Grid container spacing={2} mt={1} pl={1} pr={1}>
          <Grid item xs={12} sm={12} md={12} lg={10}>
            <Box sx={{ display: "flex", gap: "10px" }}>
              <Typography variant="h5">Processing Rule</Typography>
              <RuleEditHelp />
              <TestButton record={record} />
            </Box>
          </Grid>
          <Grid item xs={12}>
            <Box sx={{ display: "flex", gap: "20px" }}>
              <TextInput source="name" />
              <BooleanInput label="Enabled" source="enabled" />
            </Box>
          </Grid>
          <Grid item xs={12}>
            <MarkdownInput source="description" />
          </Grid>
          <Grid item xs={12}>
            <Labeled label={"Run rule after document has been:"}>
              <SelectArrayInput
                source={"trigger"}
                label={"trigger type"}
                defaultValue={"document-create"}
                required
                choices={[
                  {
                    id: "document-create",
                    name: "Created",
                  },
                  {
                    id: "document-update",
                    name: "Updated",
                  },
                ]}
              />
            </Labeled>
          </Grid>
          <Grid item xs={12}>
            <MatchTypeSelectInput source="mode" />
          </Grid>
          <Grid item xs={12}>
            <Typography variant="h5">Rule Conditions</Typography>
          </Grid>
          <Grid item xs={12}>
            <ArrayInput
              source="conditions"
              label=""
              sx={{
                background: theme.palette.background.default,
                border: "1px solid #363636FF",
                borderRadius: "5px",
                margin: "1em",
                padding: "1em",
                boxShadow: "1",
              }}
            >
              <ConditionEdit />
            </ArrayInput>
          </Grid>
          <Grid item xs={12}>
            <Typography variant="h5">Rule Actions</Typography>
            <ArrayInput
              defaultValue={{ enabled: true, on_condition: true }}
              source="actions"
              label=""
              sx={{
                background: theme.palette.background.default,
                border: "1px solid #363636FF",
                borderRadius: "5px",
                margin: "1em",
                padding: "1em",
                boxShadow: "1",
              }}
            >
              <ActionEdit />
            </ArrayInput>
          </Grid>
          <Grid item xs={12}>
            <TimestampField />
          </Grid>
          <Grid item xs={12}>
            <Toolbar>
              <SaveButton />
            </Toolbar>
          </Grid>
        </Grid>
      </Form>
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
