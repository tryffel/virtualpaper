import { FunctionField, useRecordContext } from "react-admin";
import get from "lodash/get";
import { formatRelative } from "date-fns";

export const UpdatedAtField = ({
  source,
  label,
}: {
  source: string;
  label?: string;
}) => {
  const record = useRecordContext();
  const rawValue = get(record, source);
  if (!rawValue) {
    return null;
  }

  const format = () => {
    return formatRelative(new Date(rawValue), new Date());
  };

  return <FunctionField render={format} label={label} />;
};
