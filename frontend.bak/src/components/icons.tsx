import * as Icons from "@mui/icons-material";
import React from "react";
const iconNames = Object.keys(Icons);

export const iconExists = (name: string) => {
  return name !== "" && iconNames.includes(name);
};

export const IconByName = ({
  name,
  color,
}: {
  name: string;
  color?: string;
}) => {
  const exists = iconExists(name);
  return exists
    ? // @ts-ignore
      React.createElement(Icons[name], { color })
    : null;
};
