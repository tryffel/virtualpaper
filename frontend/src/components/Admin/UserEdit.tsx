import React from "react";
import {
  BooleanInput,
  Edit,
  HttpError,
  Labeled,
  PasswordInput,
  SimpleForm,
  TextField,
  TextInput,
  useNotify,
} from "react-admin";
import { useNavigate } from "react-router-dom";

export const AdminEditUser = () => {
  const navigate = useNavigate();
  const notify = useNotify();

  const onError = (error: any) => {
    const e = error as HttpError;
    if (e.message === "authentication required") {
      notify("Authentication required", { type: "error" });
      navigate("/auth/confirm-authentication");
    }
  };

  return (
    <Edit
      redirect={"/admin"}
      mutationMode="pessimistic"
      mutationOptions={{ onError }}
    >
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
