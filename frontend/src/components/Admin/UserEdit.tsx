import React from "react";
import {
  BooleanInput,
  Edit,
  Labeled,
  PasswordInput,
  SimpleForm,
  TextField,
  TextInput,
} from "react-admin";

export const AdminEditUser = () => {
  return (
    <Edit redirect={"/admin"}>
      <SimpleForm defaultValues={{ password: "" }}>
        <Labeled label={"User id"}>
          <TextField source={"id"} />
        </Labeled>
        <Labeled label={"Username"}>
          <TextField source={"user_name"} />
        </Labeled>
        <TextInput source={"email"} />
        <BooleanInput source={"is_active"} label={"Active"} />
        <BooleanInput source={"is_admin"} label={"Administrator"} />
        <PasswordInput source={"password"} label={"Reset Password"} />
        <Labeled label={"Created at"}>
          <TextField source={"created_at"} />
        </Labeled>
        <Labeled label={"Last updated"}>
          <TextField source={"updated_at"} />
        </Labeled>
      </SimpleForm>
    </Edit>
  );
};
