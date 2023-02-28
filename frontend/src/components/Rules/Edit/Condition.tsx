import {
  BooleanInput,
  FormDataConsumer,
  ReferenceInput,
  SelectInput,
  SimpleFormIterator,
  TextInput,
} from "react-admin";
import * as React from "react";
import { Box, Grid, Tooltip, Typography } from "@mui/material";
import { CheckBoxInput } from "../../primitives/CheckBox";
import { MetadataValueInput } from "./Metadata";

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
      required
    />
  );
};

export const ConditionEdit = () => {
  return (
    <SimpleFormIterator source={"conditions"}>
      <FormDataConsumer>
        {({ scopedFormData, getSource }) => {
          return getSource ? (
            <div>
              <Grid container sx={{ flexFlow: "row" }}>
                <Grid item xs={12} md={8} lg={12}>
                  <Box flex={2}>
                    <BooleanInput
                      label="Enabled"
                      source={getSource("enabled")}
                      // @ts-ignore
                      record={scopedFormData}
                      initialValue={true}
                    />
                  </Box>
                  <Box flex={2} sx={{ pb: 2 }}>
                    <CheckBoxInput
                      label="Case insensitive"
                      source={getSource("case_insensitive")}
                      // @ts-ignore
                      record={scopedFormData}
                      defaultValue={true}
                    />
                    <CheckBoxInput
                      label="Negate"
                      source={getSource("inverted_match")}
                      // @ts-ignore
                      record={scopedFormData}
                      defaultValue={false}
                    />
                    <CheckBoxInput
                      label="Regex"
                      source={getSource("is_regex")}
                      // @ts-ignore
                      record={scopedFormData}
                      defaultValue={false}
                    />
                  </Box>
                </Grid>
                <Grid item xs={12} md={8} lg={12}>
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
                  scopedFormData.condition_type === "metadata_has_key_value" ? (
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
                </Grid>
              </Grid>
            </div>
          ) : null;
        }}
      </FormDataConsumer>
    </SimpleFormIterator>
  );
};
