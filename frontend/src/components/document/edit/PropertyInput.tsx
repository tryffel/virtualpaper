import {
  ArrayInput,
  ReferenceInput,
  SelectInput,
  SimpleFormIterator,
  TextInput,
} from "react-admin";

export const PropertyArrayInput = ({
  source,
  label,
}: {
  source: string;
  label?: string;
}) => {
  return (
    <ArrayInput source={source} label={label ?? "Properties"}>
      <SimpleFormIterator inline disableReordering fullWidth>
        <ReferenceInput
          label="Property"
          source="property_id"
          reference="properties"
          fullWidth
          sort={{ field: "name", order: "ASC" }}
        >
          <SelectInput
            optionText="name"
            data-testid="metadata-key"
            variant={"standard"}
            sx={{ mt: "3px" }}
          />
        </ReferenceInput>
        <TextInput source={"value"} />
        <TextInput source={"description"} />
      </SimpleFormIterator>
    </ArrayInput>
  );
};

export interface MetadataValueInputProps {
  source: string;
  record: any;
  label: string;
  fullWidth?: boolean;
}
