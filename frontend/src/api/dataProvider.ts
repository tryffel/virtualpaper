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

import { stringify } from "query-string";
import { HttpError } from "react-admin";
import { fetchUtils, DataProvider } from "ra-core";

import { config } from ".././env";
const apiUrl = config.url;

/* modified from ra-core fetchUtils, use json.error if any. */
const fetchJson = (url: string, options = {}) => {
  const requestHeaders = fetchUtils.createHeadersFromOptions(options);
  const token = localStorage.getItem("auth");
  requestHeaders.set("Authorization", `Bearer ${token}`);

  return fetch(url, { ...options, headers: requestHeaders })
    .then((response) =>
      response.text().then((text) => ({
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
          new HttpError((json && json.Error) || statusText, status, json)
        );
      }
      return Promise.resolve({ status, headers, body, json });
    });
};

const httpClient = (url: string, options: fetchUtils.Options = {}) => {
  if (!options.headers) {
    options.headers = new Headers({ Accept: "application/json" });
  }
  //const  token  = localStorage.getItem('auth');
  //options.headers['authorization'] = ""; //('Authorization', `Bearer ${token}`);
  return fetchJson(url, options);
};
const countHeader = "Content-Range";

export const dataProvider: DataProvider = {
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

    if (resource === "metadata/values") {
      // @ts-ignore
      if (!params.id) {
        // @ts-ignore
        params.id = 1;
      }
      // @ts-ignore
      url = `${apiUrl}/metadata/keys/${params.id}/values?${stringify(query)}`;
    } else if (resource === "admin/documents/processing") {
      url = `${apiUrl}/admin/documents/process?${stringify(query)}`;
    }

    const options =
      countHeader === "Content-Range"
        ? {
          // Chrome doesn't return `Content-Range` header if no `Range` is provided in the request.
          headers: new Headers({
            Range: `${resource}=${rangeStart}-${rangeEnd}`,
          }),
        }
        : {};

    if (resource === "documents") {
      return httpClient(url, options)
        .then(({ headers, json }) => {
          if (!headers.has(countHeader)) {
            throw new Error(
              `The ${countHeader} header is missing in the HTTP Response. The simple REST data provider expects responses for lists of resources to contain this header with the total number of results to build the pagination. If you are using CORS, did you declare ${countHeader} in the Access-Control-Expose-Headers header?`
            );
          }
          return {
            data: json,
            total:
              countHeader === "Content-Range"
                ? parseInt(
                    // @ts-ignore
                    headers.get("content-range").split("/").pop(),
                    10
                  )
                : // @ts-ignore
                  parseInt(headers.get(countHeader.toLowerCase())),
          };
        })
        .catch((error) => {
          console.log(error);
          // bad query
          if (error.status != 400) {
            throw error;
          } else {
            return {
              data: [],
              total: 0,
            };
          }
        });
    }

    return httpClient(url, options).then(({ headers, json }) => {
      if (!headers.has(countHeader)) {
        throw new Error(
          `The ${countHeader} header is missing in the HTTP Response. The simple REST data provider expects responses for lists of resources to contain this header with the total number of results to build the pagination. If you are using CORS, did you declare ${countHeader} in the Access-Control-Expose-Headers header?`
        );
      }
      return {
        data: json,
        total:
          countHeader === "Content-Range"
            ? parseInt(
              // @ts-ignore
              headers.get("content-range").split("/").pop(),
              10
            )
            : // @ts-ignore
            parseInt(headers.get(countHeader.toLowerCase())),
      };
    });
  },

  // @ts-ignore
  getOne: (resource, params) => {
    if (resource === "documents/stats") {
      return httpClient(`${apiUrl}/${resource}`).then(({ json }) => ({
        data: json,
      }));
    } else if (resource === "preferences") {
      return httpClient(`${apiUrl}/${resource}/${params.id}`).then(({ json }) => {
        // @ts-ignore
        const isAdmin = json.is_admin;
        // @ts-ignore
        localStorage.setItem("is_admin", isAdmin == true);

        // @ts-ignore
        return {
          data: { ...json, id: "user" },
        }
      })
    } else {
      if (params.id === null || params.id === "") {
        return httpClient(`${apiUrl}/${resource}`).then(({ json }) => ({
          data: { ...json, id: "1" },
        }));
      } else {
        return httpClient(`${apiUrl}/${resource}/${params.id}`).then(
          ({ json }) => ({
            data: { ...json, id: json.id ? json.id : params.id },
          })
        );
      }
    }
  },

  getMany: (resource, params) => {
    const query = {
      filter: JSON.stringify({ id: params.ids }),
    };
    let url = `${apiUrl}/${resource}?${stringify(query)}`;

    if (
      resource === "metadata/values" &&
      params.ids &&
      params.ids.length === 0
    ) {
      url = `${apiUrl}/metadata/keys/${params.ids[0]}/values?${stringify(
        query
      )}`;
    }

    if (resource === "documents") {
      return Promise.all(
        params.ids.map((id) =>
          httpClient(`${apiUrl}/${resource}/${id}`, {
            method: "GET",
          })
        )
      ).then((responses) => ({ data: responses.map(({ json }) => ({ ...json, id: json.id })) }))
    };
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
      page: page,
      page_size: perPage,
      filter: JSON.stringify({
        ...params.filter,
        [params.target]: params.id,
      }),
    };
    let url = `${apiUrl}/${resource}?${stringify(query)}`;

    if (resource === "document/jobs" && params.id) {
      url = `${apiUrl}/documents/${params.id}/jobs?${stringify(query)}`;
    } else if (resource === "metadata/values" && params.id) {
      url = `${apiUrl}/metadata/keys/${params.id}/values?${stringify(query)}`;
    } else if (resource === "documents/edithistory" && params.id) {
      url = `${apiUrl}/documents/${params.id}/history?${stringify(query)}`;
    } else if (resource === "documents/linked" && params.id) {
      url = `${apiUrl}/documents/${params.id}/linked-documents?${stringify(query)}`;
    }
    const options =
      countHeader === "Content-Range"
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
          countHeader === "Content-Range"
            ? parseInt(
              // @ts-ignore
              headers.get("content-range").split("/").pop(),
              10
            )
            : // @ts-ignore
            parseInt(headers.get(countHeader.toLowerCase())),
      };
    });
  },

  update: (resource, params) => {
    let url = `${apiUrl}/${resource}/${params.id}`;
    if (resource === "metadata/values") {
      url = `${apiUrl}/metadata/keys/${params.meta.key_id}/values/${params.data.id}`;
    }

    return httpClient(url, {
      method: "PUT",
      body: JSON.stringify(params.data),
    }).then(({ json }) => ({ data: json }));
  },

  // simple-rest doesn't handle provide an updateMany route, so we fallback to calling update n times instead
  updateMany: (resource, params) =>
    Promise.all(
      params.ids.map((id) =>
        httpClient(`${apiUrl}/${resource}/${id}`, {
          method: "PUT",
          body: JSON.stringify(params.data),
        })
      )
    ).then((responses) => ({ data: responses.map(({ json }) => json.id) })),

  // @ts-ignore
  create: (resource, params) => {
    if (resource === "documents" && params.data.file) {
      const file = params.data.file;

      let data = new FormData();
      data.append("name", file.rawFile.name);
      data.append(file.rawFile.name, file.rawFile);

      const headers = new Headers({
        Accept: "multipart/form-data",
      });

      return httpClient(`${apiUrl}/${resource}`, {
        method: "POST",
        body: data,
        headers: headers,
      }).then(({ json }) => ({
        data: { ...params.data, id: json.id },
      }));

    }
    // @ts-ignore
    if (resource === "metadata/values" && params.data) {
      return httpClient(`${apiUrl}/metadata/keys/${params.data.id}/values`, {
        method: "POST",
        body: JSON.stringify(params.data),
      }).then(({ json }) => ({
        data: json,
      }));
    } if (resource === "documents/bulkEdit") {
      return httpClient(`${apiUrl}/${resource}`, {
        method: "POST",
        body: JSON.stringify(params.data),
      }).then(({ json }) => ({
        data: { id: 'empty' },
      }));

    } else {
      return httpClient(`${apiUrl}/${resource}`, {
        method: "POST",
        body: JSON.stringify(params.data),
      }).then(({ json }) => ({
        data: { ...params.data, ...json },
      }));
    }
  },

  delete: (resource, params) => {
    let url = `${apiUrl}/${resource}/${params.id}`;
    if (resource === "metadata/values") {
      url = `${apiUrl}/metadata/keys/${params.meta.key_id}/values/${params.id}`;
    }
    return httpClient(url, {
      method: "DELETE",
    }).then(({ json }) => ({ data: json }));
  },

  // simple-rest doesn't handle filters on DELETE route, so we fallback to calling DELETE n times instead
  deleteMany: (resource, params) =>
    Promise.all(
      params.ids.map((id) =>
        httpClient(`${apiUrl}/${resource}/${id}`, {
          method: "DELETE",
        })
      )
    ).then((responses) => ({ data: responses.map(({ json }) => json.id) })),

  testRule: (resource: any, params: any) =>
    httpClient(`${apiUrl}/${resource}/${params.id}/test`, {
      method: "PUT",
      body: JSON.stringify(params.data),
    }).then(({ json }) => ({
      data: { ...params.data, ...json },
    })),

  adminRequestProcessing: (params: any) =>
    httpClient(`${apiUrl}/admin/documents/process`, {
      method: "POST",
      body: JSON.stringify(params.data),
    }).then(({ json }) => ({
      data: { ...params.data, ...json },
    })),
    
    suggestSearch: (params: any) => 
    httpClient(`${apiUrl}/documents/search/suggest`, {
      method: "POST",
      body: JSON.stringify(params.data),
    }).then(({ json }) => ({
      data: { ...json },
    })),

};

export const requestDocumentProcessing = (documentId: string) => {
  httpClient(`${apiUrl}/documents/${documentId}/process`, {
    method: "POST",
  }).then(({ json }) => ({ data: json }));
};
