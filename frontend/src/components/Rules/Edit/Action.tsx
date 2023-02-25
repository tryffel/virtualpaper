import {
  BooleanInput,
  FormDataConsumer,
  ReferenceInput,
  SelectInput,
  SimpleFormIterator,
  TextInput,
  useRecordContext,
} from "react-admin";
import { Grid } from "@mui/material";
import { CheckBoxInput } from "../../primitives/CheckBox";
import { MetadataValueInput } from "./Metadata";
import * as React from "react";

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
                    <CheckBoxInput
                      source={getSource("on_condition")}
                      label={"On condition"}
                      defaultValue={true}
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
