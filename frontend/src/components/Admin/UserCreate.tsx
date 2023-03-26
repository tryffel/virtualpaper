import React from "react";
import {
  BooleanInput,
  Create,
  PasswordInput,
  SimpleForm,
  TextInput,
} from "react-admin";

export const AdminCreateUser = () => {
  return (
    <Create>
      <SimpleForm defaultValues={{ is_active: true, is_admin: false }}>
        <TextInput source={"user_name"} label={"Username"} required />
        <TextInput source={"email"} />
        <PasswordInput source={"password"} required />
        <BooleanInput source={"is_active"} label={"Active"} />
        <BooleanInput source={"is_admin"} label={"Administrator"} />
      </SimpleForm>
    </Create>
  );
};
