import * as React from "react";
import {
  CreateBase,
  email,
  Form,
  SaveButton,
  TextInput,
  useNotify,
  useRedirect,
} from "react-admin";
import { Card, CardContent, Grid, Typography } from "@mui/material";
import { doForgotPassword } from "../../api/public";

export const ForgotPassword = () => {
  const redirect = useRedirect();
  const notify = useNotify();

  const handleSubmit = (data: any) => {
    doForgotPassword(data.email)
      .then(() => {
        notify(
          "Password reset link has been sent. Please check your email to proceed.",
          { type: "warning" }
        );
        redirect("/#");
      })
      .catch((err) => {
        notify(String(err), { type: "error" });
      });
  };

  return (
    <CreateBase disableAuthentication resource={"forgot-password"}>
      <Form
        /* @ts-ignore*/
        onSubmit={handleSubmit}
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
    <Card sx={{ mt: 6 }}>
      <CardContent sx={{ margin: 2 }}>
        <Grid container flexDirection="column">
          <Grid item sx={{ pb: 4 }}>
            <Typography variant="h4">Send password reset link</Typography>
          </Grid>
          <Grid item>
            <TextInput
              label={"Email address"}
              source={"email"}
              required
              validate={email()}
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
