import React from "react";
import {
  CreateBase,
  Form,
  HttpError,
  PasswordInput,
  SaveButton,
  TextInput,
  useLogin,
  useNotify,
} from "react-admin";
import {
  Box,
  Button,
  Card,
  CardContent,
  Grid,
  Typography,
} from "@mui/material";
import Logo from "../../layout/Logo";
import { useNavigate } from "react-router-dom";
import CircularProgress from "@mui/material/CircularProgress";

interface LoginFields {
  username: string;
  password: string;
}

const LoadingIcon = () => {
  return <CircularProgress size={18} thickness={2} />;
};

export const LoginPage = () => {
  const [loading, setLoading] = React.useState(false);
  const navigate = useNavigate();

  const login = useLogin();
  const notify = useNotify();

  const handleSubmit = (e: LoginFields) => {
    setLoading(true);
    login({ username: e.username, password: e.password })
      .then(() => {
        setLoading(false);
        navigate("/#");
      })
      .catch((error: HttpError) => {
        console.log(
          "error",
          error,
          typeof error,
          error.status,
          error.message,
          error.body
        );
        setLoading(false);
        notify("Invalid username or password", { type: "error" });
      });
  };

  return (
    <CreateBase disableAuthentication resource={"reset-password"}>
      <Form
        // @ts-ignore
        onSubmit={handleSubmit}
      >
        <div
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
          <Box sx={{ marginTop: "4em" }}>
            <Logo />
          </Box>
          <Card sx={{ mt: 1 }}>
            <CardContent sx={{ margin: "10px" }}>
              <Grid container flexDirection="column">
                <Grid item sx={{ pb: "10px" }}>
                  <Typography variant="h6">Sign in</Typography>
                </Grid>
                <Grid item>
                  <TextInput
                    label={"Username"}
                    source={"username"}
                    required
                    fullWidth
                  />
                </Grid>
                <Grid item>
                  <PasswordInput
                    label={"password"}
                    source={"password"}
                    required
                    fullWidth
                  />
                </Grid>
                <Grid item>
                  <SaveButton
                    label={loading ? "" : "Sign in"}
                    fullWidth
                    icon={loading ? <LoadingIcon /> : <></>}
                    disabled={loading}
                    sx={{ minHeight: "2.6em" }}
                  />
                </Grid>
                <Grid item sx={{ marginTop: "10px" }}>
                  <Button href={"/#/auth/forgot-password"}>
                    Forgot password?
                  </Button>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </div>
      </Form>
    </CreateBase>
  );
};
