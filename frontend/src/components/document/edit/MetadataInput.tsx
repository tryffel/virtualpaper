import {
  ArrayInput,
  AutocompleteInput,
  FormDataConsumer,
  Loading,
  ReferenceInput,
  SelectInput,
  SimpleFormIterator,
  useGetManyReference,
} from "react-admin";
import get from "lodash/get";
import { Typography } from "@mui/material";

export const MetadataArrayInput = ({
  source,
  label,
}: {
  source: string;
  label?: string;
}) => {
  return (
    <ArrayInput source={source} label={label ?? "Metadata"}>
      <SimpleFormIterator inline disableReordering fullWidth>
        <ReferenceInput
          label="Key"
          source="key_id"
          reference="metadata/keys"
          fullWidth
        >
          <SelectInput
            optionText="key"
            data-testid="metadata-key"
            variant={"standard"}
            sx={{ mt: "3px" }}
          />
        </ReferenceInput>

        <FormDataConsumer>
          {({ scopedFormData, getSource }) =>
            scopedFormData && scopedFormData.key_id ? (
              <MetadataValueInput
                source={getSource ? getSource("value_id") : ""}
                record={scopedFormData}
                label={"Value"}
              />
            ) : null
          }
        </FormDataConsumer>
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

export const MetadataValueInput = (props: MetadataValueInputProps) => {
  let keyId = 0;
  if (props.record) {
    // @ts-ignore
    keyId = get(props.record, "key_id");
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
    return (
      <AutocompleteInput
        {...props}
        choices={data}
        optionText="value"
        variant={"standard"}
        sx={{ mt: "0px" }}
      />
    );
  } else {
    return <Loading />;
  }
};
