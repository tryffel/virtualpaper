import {
  FormDataConsumer,
  SelectInput,
  TextInput,
  useRecordContext,
} from "react-admin";
import * as React from "react";
import { IconByName, iconExists } from "../../components/icons";
import { InputAdornment, Typography } from "@mui/material";
import { get } from "lodash";

export const IconSelect = ({
  displayIcon,
  source,
}: {
  displayIcon?: boolean;
  source: string;
}) => {
  const validateIcon = (value: any) => {
    if (iconExists(value)) {
      return undefined;
    }
    return "Icon is invalid. Must be Material-ui icon. Leave empty to disable";
  };

  const renderIcon = (form: any) => {
    const color = get(form, "style.color") ?? undefined;
    if (!iconExists(get(form, source))) {
      return null;
    } else {
      return <IconByName name={get(form, source)} color={color} />;
    }
  };

  return (
    <TextInput
      source={source}
      id="icon"
      label="Icon name (Material-ui)"
      validate={validateIcon}
      defaultValue={"Label"}
      placeholder={"Label"}
      InputProps={{
        endAdornment: displayIcon && (
          <FormDataConsumer>
            {({ formData }) => (
              <InputAdornment position={"end"}>
                {renderIcon(formData)}
              </InputAdornment>
            )}
          </FormDataConsumer>
        ),
      }}
    />
  );
};

export const IconColorSelect = () => {
  const record = useRecordContext();
  const choices = [
    { id: "primary", name: "primary" },
    { id: "secondary", name: "secondary" },
    { id: "info", name: "info" },
    { id: "success", name: "success" },
    { id: "warning", name: "warning" },
    { id: "error", name: "error" },
  ];

  if (record) {
    console.log("record", record);
  }

  return (
    <SelectInput choices={choices} emptyValue={""} source={"style.color"} />
  );
};
