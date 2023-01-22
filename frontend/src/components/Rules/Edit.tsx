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
  Edit,
  SimpleForm,
  TextInput,
  RadioButtonGroupInput,
  BooleanInput,
  ReferenceInput,
  SelectInput,
  ArrayInput,
  SimpleFormIterator,
  FormDataConsumer,
  Loading,
  Error,
  AutocompleteInput,
  useGetManyReference,
  useRecordContext,
  DateField,
  Button,
} from "react-admin";
import get from "lodash/get";

import { Typography, Grid, Box, useTheme } from "@mui/material";
import { HelpButton } from "../Help";
import { MarkdownInput } from "../Markdown";
import { MetadataValueInputProps } from "../Documents/Edit";
import Test from "./Test";
import TestButton from "./Test";

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

export const ConditionTypeInput = (props: any) => {
  return (
    <SelectInput
      source={props.source}
      onChange={props.onChange}
      choices={[
        { id: "name_is", name: "Name is" },
        { id: "name_starts", name: " Name starts" },
        { id: "name_contains", name: " Name contains" },

        { id: "description_is", name: " Description is" },
        { id: "description_starts", name: " Description starts" },
        { id: "description_contains", name: " Description contains" },

        { id: "content_is", name: " Text content matches" },
        { id: "content_starts", name: " Text content starts with" },
        { id: "content_contains", name: " Text content contains" },

        { id: "date_is", name: " Date is" },
        { id: "date_after", name: " Date is after" },
        { id: "date_before", name: " Date is before" },

        { id: "metadata_has_key", name: " Metadata contains" },
        { id: "metadata_has_key_value", name: " Metadata contains key-value" },
        { id: "metadata_count", name: " Metadata count equals" },
        { id: "metadata_count_less_than", name: " Metadata count less than" },
        { id: "metadata_count_more_than", name: " Metadata count more than" },
      ]}
    />
  );
};

export const ActionTypeInput = (props: any) => {
  return (
    <SelectInput
      source={props.source}
      onChange={props.onChange}
      choices={[
        { id: "name_set", name: "Set name" },
        { id: "name_append", name: "Append name" },
        { id: "description_set", name: "Set description" },
        { id: "description_append", name: "Append description" },
        { id: "metadata_add", name: "Add metadata" },
        { id: "metadata_remove", name: "Remove metadata" },
        { id: "date_set", name: "Set date" },
      ]}
    />
  );
};

export const ConditionEdit = () => {
  const record = useRecordContext();
  const theme = useTheme();
  return (
    <SimpleFormIterator source={"conditions"}>
      <FormDataConsumer>
        {({ formData, scopedFormData, getSource }) => {
          return getSource ? (
            <div>
              <Grid container>
                <Grid item xs={12} md={8} lg={12}>
                  <Box display={{ xs: "block", sm: "flex" }}>
                    <Box flex={2}>
                      <BooleanInput
                        label="Enabled"
                        source={getSource("enabled")}
                        // @ts-ignore
                        record={scopedFormData}
                        initialValue={true}
                      />
                    </Box>
                    <Box flex={2}>
                      <BooleanInput
                        label="Case insensitive"
                        source={getSource("case_insensitive")}
                        // @ts-ignore
                        record={scopedFormData}
                      />
                    </Box>
                    <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                      <BooleanInput
                        label="Inverted"
                        source={getSource("inverted")}
                        // @ts-ignore
                        record={scopedFormData}
                      />
                    </Box>
                    <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                      <BooleanInput
                        label="Regex"
                        source={getSource("is_regex")}
                        // @ts-ignore
                        record={scopedFormData}
                      />
                    </Box>
                  </Box>
                  <Box display={{ xs: "block", sm: "flex" }}>
                    <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                      <ConditionTypeInput
                        //label="Type"
                        source={getSource("condition_type")}
                        record={scopedFormData}
                      />
                    </Box>
                    {scopedFormData &&
                    scopedFormData.condition_type &&
                    (scopedFormData.condition_type.startsWith("date") ||
                      !scopedFormData.condition_type.startsWith(
                        "metadata_has_key"
                      )) ? (
                      <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                        <TextInput
                          label="Filter"
                          source={getSource("value")}
                          // @ts-ignore
                          record={scopedFormData}
                          fullWidth
                          resettable
                        />
                      </Box>
                    ) : null}
                    {scopedFormData &&
                    scopedFormData.condition_type &&
                    scopedFormData.condition_type.startsWith("date") ? (
                      <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                        <TextInput
                          label="Date format"
                          source={getSource("date_fmt")}
                          // @ts-ignore
                          record={scopedFormData}
                          fullWidth
                        />
                      </Box>
                    ) : null}
                    {scopedFormData &&
                    scopedFormData.condition_type &&
                    scopedFormData.condition_type.startsWith(
                      "metadata_has_key"
                    ) ? (
                      <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                        <ReferenceInput
                          label="Key"
                          source={getSource("metadata.key_id")}
                          record={scopedFormData}
                          reference="metadata/keys"
                          fullWidth
                        >
                          <SelectInput optionText="key" fullWidth />
                        </ReferenceInput>
                      </Box>
                    ) : null}
                    {scopedFormData &&
                    scopedFormData.condition_type &&
                    scopedFormData.condition_type ===
                      "metadata_has_key_value" ? (
                      <Box flex={2} mr={{ xs: 0, sm: "0.5em" }}>
                        <MetadataValueInput
                          source={getSource("metadata.value_id")}
                          keySource={"metadata.key_id"}
                          record={scopedFormData}
                          label={"Value"}
                          fullWidth
                        />
                      </Box>
                    ) : null}
                  </Box>
                </Grid>
              </Grid>
            </div>
          ) : null;
        }}
      </FormDataConsumer>
    </SimpleFormIterator>
  );
};

export const ActionEdit = () => {
  const record = useRecordContext();

  return record ? (
    <SimpleFormIterator source={"actions"}>
      <FormDataConsumer>
        {({ formData, scopedFormData, getSource }) => {
          return getSource ? (
            <Grid container>
              <Grid container>
                <Grid container display="flex" spacing={2}>
                  <Grid item flex={1} ml="0.5em">
                    <ActionTypeInput
                      //label="Type"
                      // @ts-ignore
                      source={getSource("action")}
                      record={scopedFormData}
                    />
                  </Grid>
                  <Grid item flex={1} ml="0.5em">
                    <BooleanInput
                      label="Enabled"
                      source={getSource("enabled")}
                      // @ts-ignore
                      record={scopedFormData}
                    />
                  </Grid>
                  <Grid item flex={1} mr="0.5em">
                    <BooleanInput
                      label="On condition"
                      source={getSource("on_condition")}
                    />
                  </Grid>
                </Grid>
                {scopedFormData &&
                scopedFormData.action &&
                !scopedFormData.action.startsWith("metadata") ? (
                  <Grid container display="flex" spacing={2}>
                    <Grid item flex={1} ml="0.5em">
                      <TextInput label="Value" source={getSource("value")} />
                    </Grid>
                  </Grid>
                ) : null}
                {scopedFormData &&
                scopedFormData.action &&
                scopedFormData.action.startsWith("metadata") ? (
                  <Grid container display="flex" spacing={2}>
                    <Grid item xs={8} md={4} lg={3}>
                      <ReferenceInput
                        label="Key"
                        source={getSource("metadata.key_id")}
                        record={scopedFormData}
                        reference="metadata/keys"
                        fullWidth
                      >
                        <SelectInput optionText="key" fullWidth />
                      </ReferenceInput>
                    </Grid>
                    <Grid item xs={8} md={4} lg={3}>
                      <MetadataValueInput
                        source={getSource("metadata.value_id")}
                        keySource={"metadata.key_id"}
                        record={scopedFormData}
                        label={"Value"}
                        fullWidth
                      />
                    </Grid>
                  </Grid>
                ) : null}
              </Grid>
            </Grid>
          ) : null;
        }}
      </FormDataConsumer>
    </SimpleFormIterator>
  ) : null;
};

interface InputProps extends MetadataValueInputProps {
  keySource: string;
}

const MetadataValueInput = (props: InputProps) => {
  let keyId = 0;
  if (props.record) {
    keyId = get(props.record, props.keySource);
  }

  const { data, isLoading, error } = useGetManyReference("metadata/values", {
    target: "id",
    id: keyId !== 0 ? keyId : -1,
    pagination: { page: 1, perPage: 500 },
    sort: {
      field: "value",
      order: "ASC",
    },
  });

  if (!props.record) {
    return null;
  }

  if (isLoading) return <Loading />;
  if (error) return <Typography>Error {error.message}</Typography>;

  if (data) {
    return <AutocompleteInput {...props} choices={data} optionText="value" />;
  } else {
    return <Loading />;
  }
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

const RuleEditHelp = () => {
  const dateRegexExample = `(\d{4}-\d{1,2}-\d{1,2})`;

  return (
    <HelpButton title="Edit Rule">
      <Typography variant="h6" color="textPrimary">
        What are processing rules
      </Typography>
      <p>
        Processing rules are a set of instructions that try to minimize the
        manual work when uploading documents by automatically detecting document
        and modifying its contents. The idea is to configure a set of conditions
        that must match, and running a set of actions for documents that match
        the conditions.
      </p>
      <p>
        Processing rules can try to detect document content, name, description
        or metadata and modify the same fields.
      </p>
      <p>
        If e.g. Google regularly sends invoices, it might make sense to try to
        detect documents where:
        <ol>
          <li>Sender is Google</li>
          AND
          <li>Document is an invoice</li>
        </ol>
        If so, then:
        <ol>
          <li>Set name to 'Google monthly invoice'</li>
          <li>Add metadata 'class:invoice'</li>
          <li>Add metadata 'company:google'</li>
        </ol>
      </p>
      <Typography variant="h5" color="textPrimary">
        Instructions
      </Typography>
      <p>
        Match conditions:
        <ul>
          <li>Match all: all conditions must match</li>
          <li>Match any: any condition must match</li>
        </ul>
      </p>
      <p>
        Condition settings
        <ul>
          <li>
            Enabled: user can toggle each condition on and off without deleting
            it
          </li>
          <li>
            Case insensitive: whether to match text in case insensitive. Only
            applies to name, description or content filters
          </li>
          <li>
            Inverted: boolean negation. If selected, the condition result is
            negated.
          </li>
          <li>
            Regex: whether or not the filter is a regular expression. Only
            applies to name, description or content filters.
          </li>
        </ul>
      </p>
      <p>
        Action settings
        <ul>
          <li>
            Enabled: user can toggle each action on and off without deleting it.
          </li>
          <li>
            On condition: if true, action is only executed when conditions are
            met. If false, action is executed if conditions are not met
          </li>
        </ul>
      </p>
      <Typography variant="h6">Extracting date</Typography>
      <p>
        Matching a date from the document is a special case of condition. By
        setting condition type to 'date is' the automation searches for dates
        inside the document. The automation searches for the given regular
        expressions to match date. If date is found, then the date time is
        extracted using the date format. Thus regular expression controls
        finding the date time text inside the document and date format controls
        how the matched date string is converted to date time. For more info on
        possible time formats, see Golang's documentation on time formats:
        <a href="https://pkg.go.dev/time#pkg-constants">pkg.go.dev/time</a>
      </p>
      <p>
        Fox configuring date extraction set following settings:
        <ol>
          <li>Set regex to true</li>
          <li>Enter regular expression</li>
          <li>
            Enter a valid date time format as per Golang time parsing formats.{" "}
          </li>
        </ol>
        Example values could be:
        <ol>
          <li>filter: '{dateRegexExample}' would match date 2022-07-15</li>
          <li>Date format would thus be '2006-1-2'</li>
        </ol>
      </p>
      In this case the user must set the 'filter' as a valid regular expression
      to match the date. E.g.
      <Typography variant="h6" color="textPrimary">
        Tips
      </Typography>
      <ul>
        <li>
          Try to create filters that are as strict as possible. E.g. matching
          content with 'Google' probably matches many more documents than
          intended. Specific email address or bank account might limit the
          results down.
        </li>
      </ul>
    </HelpButton>
  );
};
