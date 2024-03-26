import { config } from "../env";

export const doResetPassword = (
  token: string,
  tokenId: string,
  password: string
) => {
  const request = new Request(config.url + "/auth/reset-password", {
    method: "POST",
    body: JSON.stringify({
      password,
      id: Number(tokenId),
      token,
    }),
    headers: new Headers({ "Content-Type": "application/json" }),
  });

  return fetch(request)
    .then((response) => {
      if (response.status < 200 || response.status >= 300) {
        throw new Error(response.statusText);
      }
      return response.json();
    })
    .then((auth) => {});
};

export const doForgotPassword = (email: string) => {
  const request = new Request(config.url + "/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify({
      email,
    }),
    headers: new Headers({ "Content-Type": "application/json" }),
  });

  return fetch(request)
    .then((response) => {
      if (response.status < 200 || response.status >= 300) {
        // @ts-ignore
        throw new Error(response.statusText);
      }
      return response.json();
    })
    .then((auth) => {});
};
