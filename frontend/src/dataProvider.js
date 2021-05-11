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

import { stringify } from 'query-string';
import { HttpError } from 'react-admin';

import { config } from "./env.js";

const apiUrl = config.url;


/* copied from ra-core fetchUtils. */
export const createHeadersFromOptions = (options) => {
    const requestHeaders = (options.headers ||
        new Headers({
            Accept: 'application/json',
        }));
    if (
        !requestHeaders.has('Content-Type') &&
        !(options && (!options.method || options.method === 'GET')) &&
        !(options && options.body && options.body instanceof FormData)
    ) {
        requestHeaders.set('Content-Type', 'application/json');
    }
    if (options.user && options.user.authenticated && options.user.token) {
        requestHeaders.set('Authorization', options.user.token);
    }
    return requestHeaders;
};


/* modified from ra-core fetchUtils, use json.error if any. */
const fetchJson = (url, options= {}) => {
    const requestHeaders = createHeadersFromOptions(options);

    return fetch(url, { ...options, headers: requestHeaders })
        .then(response =>
            response.text().then(text => ({
                status: response.status,
                statusText: response.statusText,
                headers: response.headers,
                body: text,
            }))
        )
        .then(({ status, statusText, headers, body }) => {
            let json;
            try {
                json = JSON.parse(body);
            } catch (e) {
                // not json, no big deal
            }
            if (status < 200 || status >= 300) {
                return Promise.reject(
                    new HttpError(
                        (json && json.Error) || statusText,
                        status,
                        json
                    )
                );
            }
            return Promise.resolve({ status, headers, body, json });
        });
};

const httpClient = (url, options = {}) => {
    if (!options.headers) {
        options.headers = new Headers({ Accept: 'application/json' });
    }
    const  token  = localStorage.getItem('auth');
    options.headers.set('Authorization', `Bearer ${token}`);
    return fetchJson(url, options);
};
const countHeader = "Content-Range";

export const dataProvider = {
    getList: (resource, params) => {
        const { page, perPage } = params.pagination;
        const { field, order } = params.sort;

        const rangeStart = (page - 1) * perPage;
        const rangeEnd = page * perPage - 1;

        const query = {
            sort: JSON.stringify([field, order]),
            page: page,
            page_size: perPage,
            filter: JSON.stringify(params.filter),
        };
        let url = `${apiUrl}/${resource}?${stringify(query)}`;

        if (resource === 'metadata/values') {
            if (!params.id) {
                params.id = 1;
            }
            url = `${apiUrl}/metadata/keys/${params.id}/values?${stringify(query)}`;
        }

        const options =
            countHeader === 'Content-Range'
                ? {
                    // Chrome doesn't return `Content-Range` header if no `Range` is provided in the request.
                    headers: new Headers({
                        Range: `${resource}=${rangeStart}-${rangeEnd}`,
                    }),
                }
                : {};

        return httpClient(url, options).then(({ headers, json }) => {
            if (!headers.has(countHeader)) {
                throw new Error(
                    `The ${countHeader} header is missing in the HTTP Response. The simple REST data provider expects responses for lists of resources to contain this header with the total number of results to build the pagination. If you are using CORS, did you declare ${countHeader} in the Access-Control-Expose-Headers header?`
                );
            }
            return {
                data: json,
                total:
                    countHeader === 'Content-Range'
                        ? parseInt(
                        headers.get('content-range').split('/').pop(),
                        10
                        )
                        : parseInt(headers.get(countHeader.toLowerCase())),
            };
        });
    },

    getOne: (resource, params) => {

        if (resource === "documents/stats") {
            return httpClient(`${apiUrl}/${resource}`).then(({ json }) => ({

                data: json,
        }))}
        else if (resource === "preferences") {
            return httpClient(`${apiUrl}/${resource}/${params.id}`).then(({ json }) => ({
            data: {...json, id: 'user'}
            }))}
        else {
            if (params.id === null || params.id === "") {
                return httpClient(`${apiUrl}/${resource}`).then(({json}) => ({
                    data: {...json, id: "1"},
                }))
            } else {
                return httpClient(`${apiUrl}/${resource}/${params.id}`).then(({json}) => ({
                    data: {...json, id: json.id ? json.id : params.id},
                }))
            }}},

    getMany: (resource, params) => {
        const query = {
            filter: JSON.stringify({ id: params.ids }),
        };
        let url = `${apiUrl}/${resource}?${stringify(query)}`;

        if (resource === 'metadata/values' && params.ids && params.ids.length === 0) {
            url = `${apiUrl}/metadata/keys/${params.ids[0]}/values?${stringify(query)}`;
        }

        return httpClient(url).then(({ json }) => ({ data: json }));
    },

    getManyReference: (resource, params) => {
        const { page, perPage } = params.pagination;
        const { field, order } = params.sort;

        const rangeStart = (page - 1) * perPage;
        const rangeEnd = page * perPage - 1;

        const query = {
            sort: JSON.stringify([field, order]),
            range: JSON.stringify([(page - 1) * perPage, page * perPage - 1]),
            filter: JSON.stringify({
                ...params.filter,
                [params.target]: params.id,
            }),
        };
        let url = `${apiUrl}/${resource}?${stringify(query)}`;
        if (resource !== 'metadata/values' || params.id) {
            url = `${apiUrl}/metadata/keys/${params.id}/values?${stringify(query)}`;
        }
        const options =
            countHeader === 'Content-Range'
                ? {
                    // Chrome doesn't return `Content-Range` header if no `Range` is provided in the request.
                    headers: new Headers({
                        Range: `${resource}=${rangeStart}-${rangeEnd}`,
                    }),
                }
        : {};

        return httpClient(url, options).then(({ headers, json }) => {
            if (!headers.has(countHeader)) {
                throw new Error(
                    `The ${countHeader} header is missing in the HTTP Response. The simple REST data provider expects responses for lists of resources to contain this header with the total number of results to build the pagination. If you are using CORS, did you declare ${countHeader} in the Access-Control-Expose-Headers header?`
                );
            }
            return {
                data: json,
                total:
                    countHeader === 'Content-Range'
                        ? parseInt(
                        headers.get('content-range').split('/').pop(),
                        10
                        )
                        : parseInt(headers.get(countHeader.toLowerCase())),
            };
        });
    },

    update: (resource, params) =>  {
        let url = `${apiUrl}/${resource}/${params.id}`;

        if (resource === 'metadata/values') {
            url =`${apiUrl}/metadata/keys/${params.key_id}/values/${params.data.id}`;
        }

        return httpClient(url, {
            method: 'PUT',
            body: JSON.stringify(params.data),
        }).then(({ json }) => ({ data: json }))},

    // simple-rest doesn't handle provide an updateMany route, so we fallback to calling update n times instead
    updateMany: (resource, params) =>
        Promise.all(
            params.ids.map(id =>
                httpClient(`${apiUrl}/${resource}/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify(params.data),
                })
            )
        ).then(responses => ({ data: responses.map(({ json }) => json.id) })),

    create: (resource, params ) => {
        if (resource === 'documents' && params.data.file) {
            const file = params.data.file;

            let data = new FormData();
            data.append("name", file.name)
            data.append("file", file.rawFile);

            const headers = new Headers({
                Accept: "multipart/form-data"
            });

            return httpClient(`${apiUrl}/${resource}`, {
                method: 'POST',
                body: data,
                headers: headers
            }).then(({json}) => ({
                data: {...params.data, id: json.id},
            }))

        } if (resource === 'metadata/values' && params.key_id) {
            return httpClient(`${apiUrl}/metadata/keys/${params.key_id}/values`, {
                method: 'POST',
                body: JSON.stringify(params.data),
            }).then(({ json }) => ({
                data: json,
            }))
        } else {
            httpClient(`${apiUrl}/${resource}`, {
                method: 'POST',
                body: JSON.stringify(params.data),
            }).then(({json}) => ({
                data: {...params.data, id: json.id},
            }))
        }
    },

    delete: (resource, params) =>
        httpClient(`${apiUrl}/${resource}/${params.id}`, {
            method: 'DELETE',
        }).then(({ json }) => ({ data: json })),

    // simple-rest doesn't handle filters on DELETE route, so we fallback to calling DELETE n times instead
    deleteMany: (resource, params) =>
        Promise.all(
            params.ids.map(id =>
                httpClient(`${apiUrl}/${resource}/${id}`, {
                    method: 'DELETE',
                })
            )
        ).then(responses => ({ data: responses.map(({ json }) => json.id) })),
};

export const requestDocumentProcessing = (documentId) => {
    httpClient(`${apiUrl}/documents/${documentId}/process`, {
        method: 'POST',
    }).then(({ json }) => ({ data: json }))
}


