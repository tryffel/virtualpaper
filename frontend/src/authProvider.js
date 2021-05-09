/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
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

import { config } from "./env.js";

const authProvider = {
    login: ({ username, password }) =>  {
        const request = new Request(config.url + '/auth/login', {
            method: 'POST',
            body: JSON.stringify({ "Username": username, "Password": password }),
            headers: new Headers({ 'Content-Type': 'application/json' }),
        });

        console.log(username);
        return fetch(request)
            .then(response => {
                if (response.status < 200 || response.status >= 300) {
                    throw new Error(response.statusText);
                }
                return response.json();
            })
            .then(auth => {
                const token = auth["Token"];
                const userId = auth["UserId"];
                localStorage.setItem('auth', token);
                localStorage.setItem('userId', userId);
            });
    },
    logout: () => {
        localStorage.removeItem('auth');
        return Promise.resolve();
    },
    checkAuth: () => localStorage.getItem('auth')
        ? Promise.resolve()
        : Promise.reject(),
    checkError: error => Promise.resolve(),
    getPermissions: () => {
        const isAdmin = localStorage.getItem('is_admin');
        if (isAdmin === null) {
            return Promise.reject();
        }
        const permissions = {
            admin: isAdmin,
        }
        return Promise.resolve(permissions);
    },
    getIdentity: () => Promise.resolve(),
};



export default authProvider;