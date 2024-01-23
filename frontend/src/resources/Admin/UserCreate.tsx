import React from "react";
import {
  BooleanInput,
  Create,
  HttpError,
  PasswordInput,
  SimpleForm,
  TextInput,
  useNotify,
} from "react-admin";
import { useNavigate } from "react-router-dom";

export const AdminCreateUser = () => {
  const navigate = useNavigate();
  const notify = useNotify();

  const onError = (error: any) => {
    const e = error as HttpError;
    if (e.message === "authentication required") {
      notify("Authentication required", { type: "error" });
      navigate("/auth/confirm-authentication");
    } else if (e.status === 304) {
      notify("User already exists", { type: "error" });
    } else {
      notify(e.message, { type: "error" });
    }
  };

  return (
    <Create mutationOptions={{ onError }}>
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
