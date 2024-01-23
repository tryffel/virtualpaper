import get from "lodash/get";
import { AutocompleteInput, Loading, useGetManyReference } from "react-admin";
import { Typography } from "@mui/material";
import { MetadataValueInputProps } from "../../Documents/Edit";

interface InputProps extends MetadataValueInputProps {
  keySource: string;
}

export const MetadataValueInput = (props: InputProps) => {
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

  if (isLoading)
    return <AutocompleteInput disabled choices={[]} source={props.source} />;
  if (error) return <Typography>Error {error.message}</Typography>;

  if (data) {
    return <AutocompleteInput {...props} choices={data} optionText="value" />;
  } else {
    return <Loading />;
  }
};
