import { config } from "../env";

export const doResetPassword = (token: string, password: string) => {
  const request = new Request(config.url + "/auth/reset-password", {
    method: "POST",
    body: JSON.stringify({
      password,
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
