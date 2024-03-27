import Grid from "@mui/material/Grid";
import { DateField, Labeled } from "react-admin";
import { UpdatedAtField } from "./UpdatedAtField";

export type TimestampFieldProps = {
  createdAtSource?: string;
  updatedAtSource?: string;
};

export const TimestampField = (props: TimestampFieldProps) => {
  return (
    <Grid container paddingTop={2}>
      <Grid item xs={6} mb={1}>
        <Labeled label="Created at">
          <DateField source={props.createdAtSource ?? "created_at"} />
        </Labeled>
      </Grid>
      <Grid item xs={6} mb={1}>
        <Labeled label="Updated at">
          <UpdatedAtField source={props.updatedAtSource ?? "updated_at"} />
        </Labeled>
      </Grid>
    </Grid>
  );
};
