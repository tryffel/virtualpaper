import * as React from "react";
import {
  CreateBase,
  Form,
  PasswordInput,
  SaveButton,
  useDataProvider,
  useNotify,
} from "react-admin";
import { Button, Card, CardContent, Grid, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import CancelIcon from "@mui/icons-material/Cancel";

export const ConfirmAuthentication = () => {
  const notify = useNotify();
  const dataProvider = useDataProvider();
  const navigate = useNavigate();

  const handleSubmit = (data: any) => {
    dataProvider
      .confirmAuthentication({ data })
      .then(() => {
        notify("User confirmed", { type: "success" });
        navigate(-1);
      })
      .catch((err: any) => {
        notify(String(err), { type: "error" });
      });
  };

  const handleCancel = () => {
    localStorage.setItem("requires_reauthentication", "false");
    navigate(-1);
  };

  return (
    <CreateBase resource={"reauthenticate"}>
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
          <ResetForm onCancel={handleCancel} />
        </Grid>
      </Form>
    </CreateBase>
  );
};

const ResetForm = (props: { onCancel: () => void }) => {
  return (
    <Card sx={{ mt: 6 }}>
      <CardContent sx={{ margin: 2 }}>
        <Grid container flexDirection="column">
          <Grid item sx={{ pb: 4 }}>
            <Typography variant="h4">Authentication required</Typography>
            <Typography variant="body2">
              Please enter your password to continue.
            </Typography>
          </Grid>
          <Grid item>
            <PasswordInput
              label={"Password"}
              source={"password"}
              required
              fullWidth
            />
          </Grid>
          <Grid
            item
            container
            flexDirection={"row"}
            justifyContent={"space-evenly"}
          >
            <SaveButton label={"Confirm"} />
            <Button startIcon={<CancelIcon />} onClick={props.onCancel}>
              Cancel
            </Button>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
};
