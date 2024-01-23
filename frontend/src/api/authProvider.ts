/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import { AuthProvider, HttpError, UserIdentity } from "react-admin";
import { config } from "../env";
import { httpClient } from "./dataProvider";

const apiUrl = config.url;

interface LoginFields {
  username: string;
  password: string;
}

export interface AuthPermissions {
  admin: boolean;
  requiresReauthentication: boolean;
}

const authProvider: AuthProvider = {
  login: (login: LoginFields) => {
    const request = new Request(config.url + "/auth/login", {
      method: "POST",
      body: JSON.stringify({
        Username: login.username,
        Password: login.password,
      }),
      headers: new Headers({ "Content-Type": "application/json" }),
    });

    console.log("Logging in as: ", login.username);
    return fetch(request)
      .then((response) => {
        if (response.status < 200 || response.status >= 300) {
          throw new Error(response.statusText);
        }
        return response.json();
      })
      .then((auth) => {
        const token = auth["Token"];
        const userId = auth["UserId"];
        localStorage.setItem("auth", token);
        localStorage.setItem("userId", userId);
      });
  },
  logout: () =>
    httpClient(`${apiUrl}/auth/logout`, { method: "POST" })
      .catch(() => {})
      .then(() => {
        localStorage.removeItem("auth");
        localStorage.removeItem("is_admin");
        localStorage.removeItem("userId");
      }),
  checkAuth: () =>
    localStorage.getItem("auth") ? Promise.resolve() : Promise.reject(),
  checkError: (error: HttpError) => {
    if (error.status === 401) {
      if (error.message === "invalid token") {
        return Promise.reject();
      } else if (error.message === "authentication required") {
        localStorage.setItem("requires_reauthentication", "true");
        return Promise.resolve();
      }
      return Promise.reject();
    }
    return Promise.resolve();
  },
  getPermissions: () => {
    const isAdmin = localStorage.getItem("is_admin");
    const authRequired = localStorage.getItem("requires_reauthentication");
    const requiresReauthentication = authRequired
      ? authRequired === "true" && true
      : false;
    const permissions = {
      admin: isAdmin,
      requiresReauthentication,
    };
    return Promise.resolve(permissions);
  },
  getIdentity: (): Promise<UserIdentity> => {
    const user = localStorage.getItem("userId");
    return Promise.resolve({
      id: user ? user : "",
      fullName: "",
    });
  },
  isAdmin: () => {
    const admin = localStorage.getItem("is_admin");
    return admin === "true";
  },
  getToken: () => {
    return localStorage.getItem("auth");
  },
};

export default authProvider;
