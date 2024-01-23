import * as React from "react";
import CheckBox from "@mui/material/Checkbox";
import FormControlLabel from "@mui/material/FormControlLabel";
import FormGroup from "@mui/material/FormGroup";
import { FieldTitle } from "react-admin";
import { useInput } from "ra-core";
import { useCallback } from "react";

export const CheckBoxInput = (props: {
  source: string;
  defaultValue?: boolean;
  label: string;
}) => {
  const { source, defaultValue, label } = props;

  const {
    id,
    field,
    isRequired,
    fieldState: { error, invalid, isTouched },
    formState: { isSubmitted },
  } = useInput({
    defaultValue,
    source,
    type: "checkbox",
  });

  const handleChange = useCallback(
    (event: object) => {
      field.onChange(event);
      // Ensure field is considered as touched
      field.onBlur();
    },
    [field]
  );

  return (
    <FormGroup>
      <FormControlLabel
        control={
          <CheckBox
            checked={!!field.value}
            onChange={handleChange}
            required={isRequired}
          />
        }
        label={
          <FieldTitle label={label} source={source} isRequired={isRequired} />
        }
      />
    </FormGroup>
  );
};
