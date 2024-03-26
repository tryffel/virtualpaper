import React, { Suspense } from "react";
import { FormDataConsumer, TextInput } from "react-admin";
const IconByName = React.lazy(() => import("../../components/icons.tsx"));

import { InputAdornment } from "@mui/material";
import get from "lodash/get";

const iconExists = (iconName: string) => {
  const icon = (
    <Suspense fallback={null}>
      <IconByName name={iconName} />
    </Suspense>
  );
  return icon !== null;
};

export const IconSelect = ({
  displayIcon,
  source,
}: {
  displayIcon?: boolean;
  source: string;
}) => {
  const validateIcon = (value: string) => {
    if (iconExists(value)) {
      return undefined;
    }
    return "Icon is invalid. Must be Material-ui icon. Leave empty to disable";
  };

  const renderIcon = (form: object) => {
    const color = get(form, "style.color") ?? undefined;
    if (!iconExists(get(form, source))) {
      return null;
    } else {
      return (
        <Suspense fallback={null}>
          <IconByName name={get(form, source)} color={color} />
        </Suspense>
      );
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

export default IconSelect;
