import * as React from "react";
import {
  CreateBase,
  Form,
  PasswordInput,
  SaveButton,
  useNotify,
  useRedirect,
  ValidateForm,
} from "react-admin";
import { Card, CardContent, Grid, Typography } from "@mui/material";
import { useSearchParams } from "react-router-dom";
import { doResetPassword } from "../../api/public";

interface PasswordFields {
  password?: string;
  password_confirmation?: string;
}

export const ResetPassword = () => {
  const [params] = useSearchParams();
  const redirect = useRedirect();
  const notify = useNotify();
  const token = params.get("token");
  const id = params.get("id");

  if (!token || !id) {
    notify("Reset link seems invalid. Please create a new link.", {
      type: "error",
    });
    redirect("/#");
  }

  const validateForm = (values: ValidateForm) => {
    const form = values as unknown as PasswordFields;
    if (!form.password || !form.password_confirmation) {
      return { error: "error" };
    }

    if (form.password?.length <= 8) {
      return { password: "Minimum of 8 characters is required" };
    }

    if (form.password !== form.password_confirmation) {
      return { password_confirmation: "passwords don't match" };
    }
    return {};
  };

  const handleSubmit = (data: PasswordFields) => {
    doResetPassword(token as string, id as string, data.password!)
      .then(() => {
        notify("Password has been reset. Please login to continue.", {
          type: "warning",
        });
        redirect("/#");
      })
      .catch((err) => {
        notify(String(err), { type: "error" });
      });
  };

  return (
    <CreateBase disableAuthentication resource={"reset-password"}>
      <Form
        /* @ts-ignore*/
        validate={validateForm}
        /* @ts-ignore*/
        onSubmit={handleSubmit}
        //defaultValues={{ password: "", password_confirmation: "" }}
      >
        <Grid
          container
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "flex-start",
            backgroundImage:
              "radial-gradient(circle at 50% 14em, #313264 0%, #00023b 60%, #00023b 100%)",
            height: "100vh",
          }}
        >
          <ResetForm />
        </Grid>
      </Form>
    </CreateBase>
  );
};

const ResetForm = () => {
  return (
    <Card sx={{ mt: 8 }}>
      <CardContent sx={{ margin: 2 }}>
        <Grid container flexDirection="column">
          <Grid item sx={{ pb: 4 }}>
            <Typography variant="h4">Reset password</Typography>
          </Grid>
          <Grid item>
            <PasswordInput
              label={"new password"}
              source={"password"}
              required
            />
          </Grid>
          <Grid item>
            <PasswordInput
              label={"confirm password"}
              source={"password_confirmation"}
              required
            />
          </Grid>
          <Grid item>
            <SaveButton label={"Reset"} fullWidth />
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
};
