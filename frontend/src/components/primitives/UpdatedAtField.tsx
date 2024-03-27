import { FunctionField, useRecordContext } from "react-admin";
import get from "lodash/get";
import { prettifyRelativeTime } from "../util.ts";
import { Tooltip, Typography } from "@mui/material";
import React from "react";

export const UpdatedAtField = ({
  source,
  label,
}: {
  source: string;
  label?: string;
}) => {
  const record = useRecordContext();
  const rawValue = get(record, source);
  const fullTime = React.useMemo(() => {
    if (!rawValue) {
      return "";
    }
    const date = new Date(rawValue);
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
  }, [rawValue]);

  if (!rawValue) {
    return null;
  }

  const format = () => {
    return (
      <Tooltip title={fullTime}>
        <Typography variant={"body2"}>
          {prettifyRelativeTime(rawValue)}
        </Typography>
      </Tooltip>
    );
  };

  return <FunctionField render={format} label={label} />;
};
