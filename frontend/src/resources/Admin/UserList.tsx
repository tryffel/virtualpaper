import {
  BooleanField,
  Button,
  Datagrid,
  DateField,
  EditButton,
  Labeled,
  ListContextProvider,
  TextField,
  useListController,
  useRecordContext,
} from "react-admin";
import { ByteToString } from "../../components/util";
import { Grid, Typography, useMediaQuery } from "@mui/material";
import { BooleanIndexingStatusField } from "../../components/IndexingStatus";
import * as React from "react";
import { useNavigate } from "react-router-dom";
import AddIcon from "@mui/icons-material/Add";

export const AdminShowUsers = (props: any) => {
  const listContext = useListController({ ...props, resource: "admin/users" });
  const isSmall = useMediaQuery((theme: any) => theme.breakpoints.down("sm"));

  return (
    <ListContextProvider value={listContext}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography variant={"h6"}>
            Total users: {listContext.total}
          </Typography>
        </Grid>

        <Grid item xs={12}>
          <Datagrid expand={ShowExpandedUser}>
            <TextField source="id"></TextField>
            <TextField source="user_name"></TextField>
            {!isSmall ? <TextField source="email"></TextField> : null}
            <BooleanField source="is_active"></BooleanField>
          </Datagrid>
        </Grid>
        <Grid item xs={12}>
          <AddUserButton />
        </Grid>
      </Grid>
    </ListContextProvider>
  );
};
const ShowExpandedUser = () => {
  const record = useRecordContext();
  const documentsSize = record ? ByteToString(record.documents_size) : "0";

  return (
    <Grid container spacing={3}>
      <Grid item xs={4} sm={3}>
        <BooleanIndexingStatusField source="indexing" />
      </Grid>
      <Grid item xs={4} sm={3} md={2} lg={1}>
        <Labeled label={"Created at"}>
          <DateField source="created_at" />
        </Labeled>
      </Grid>
      <Grid item xs={4} sm={3} md={2} lg={1}>
        <Labeled label={"Updated at"}>
          <DateField source="updated_at" />
        </Labeled>
      </Grid>
      <Grid item xs={4} sm={3} md={2} lg={1}>
        <Labeled label={"Administrator"}>
          <BooleanField source="is_admin" />
        </Labeled>
      </Grid>
      <Grid item xs={4} sm={3} md={2} lg={1}>
        <Labeled label={"# documents"}>
          <TextField source="documents_count" />
        </Labeled>
      </Grid>
      <Grid item xs={4} sm={3} md={2} lg={1}>
        <Labeled label={"Storage size"}>
          <Typography variant="body1">{documentsSize}</Typography>
        </Labeled>
      </Grid>
      <Grid item xs={2} sm={1} marginLeft={"auto"}>
        <EditButton resource={"admin/users"} />
      </Grid>
    </Grid>
  );
};

const AddUserButton = () => {
  const navigate = useNavigate();
  return (
    <Button onClick={() => navigate("/admin/users/create")} label={"Add user"}>
      <AddIcon />
    </Button>
  );
};
