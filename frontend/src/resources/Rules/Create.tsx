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
  Create,
  SimpleForm,
  TextInput,
  RadioButtonGroupInput,
  BooleanInput,
} from "react-admin";
import { Typography } from "@mui/material";
import { MarkdownInput } from "../../components/markdown";

const defaultValue = {
  name: "",
  description: "",
  enabled: true,
  mode: "match_all",
  conditions: [
    {
      enabled: true,
      condition_type: "content_contains",
      value: "empty",
    },
  ],
  actions: [
    {
      enabled: true,
      action: "name_append",
      value: "",
    },
  ],
};
export const RuleCreate = () => {
  return (
    <Create title={"Create rule"} redirect="edit">
      <SimpleForm defaultValues={defaultValue}>
        <Typography variant="h5">Processing Rule</Typography>
        <BooleanInput label="Enabled" source="enabled" />
        <TextInput source="name" fullWidth={true} />
        <MarkdownInput source="description" />
        <RadioButtonGroupInput
          label="Match conditions"
          source="mode"
          fullWidth={true}
          choices={[
            { id: "match_all", name: "Match all" },
            { id: "match_any", name: "Match any" },
          ]}
        />
      </SimpleForm>
    </Create>
  );
};
