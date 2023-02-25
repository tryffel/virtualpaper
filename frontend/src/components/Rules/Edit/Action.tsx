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
      required
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
            <ScopedActionEdit
              scopedFormData={scopedFormData}
              getSource={getSource}
            />
          ) : null;
        }}
      </FormDataConsumer>
    </SimpleFormIterator>
  ) : null;
};

const ScopedActionEdit = (props: { scopedFormData: any; getSource: any }) => {
  const { scopedFormData, getSource } = props;

  const hasAction = !!scopedFormData.action;
  const editingMetadata = scopedFormData?.action?.startsWith("metadata");

  return (
    <Grid
      container
      lg={12}
      md={10}
      direction="row"
      justifyContent="space-evenly"
      alignItems="center"
      sx={{ mt: 0.5 }}
    >
      <Grid
        container
        xs={8}
        sm={8}
        md={3}
        alignItems={"flex-start"}
        justifyContent={"flex-start"}
      >
        <Grid item>
          <BooleanInput
            label="Enabled"
            source={getSource("enabled")}
            // @ts-ignore
            record={scopedFormData}
          />
        </Grid>
        <Grid item>
          <CheckBoxInput
            source={getSource("on_condition")}
            label={"On condition"}
            defaultValue={true}
          />
        </Grid>
      </Grid>

      <Grid item sm={6} md={8} lg={8}>
        <Grid container>
          <Grid item xs={12} sm={12} md={6} sx={{ pr: 1 }}>
            <ActionTypeInput
              //label="Type"
              // @ts-ignore
              source={getSource("action")}
              record={scopedFormData}
            />
          </Grid>
          <Grid item xs={12} sm={12} md={6} lg={6}>
            {hasAction && !editingMetadata && (
              <Grid item sm={12}>
                <TextInput label="Value" source={getSource("value")} />
              </Grid>
            )}
            {hasAction && editingMetadata && (
              <Grid item sm={12} md={6} lg={6}>
                <ReferenceInput
                  label="Key"
                  source={getSource("metadata.key_id")}
                  record={scopedFormData}
                  reference="metadata/keys"
                  fullWidth
                >
                  <SelectInput optionText="key" fullWidth />
                </ReferenceInput>
                <MetadataValueInput
                  source={getSource("metadata.value_id")}
                  keySource={"metadata.key_id"}
                  record={scopedFormData}
                  label={"Value"}
                  fullWidth
                />
              </Grid>
            )}
          </Grid>
        </Grid>
      </Grid>
    </Grid>
  );
};
