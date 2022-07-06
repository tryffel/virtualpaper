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

import * as React from 'react';
import { useState } from 'react';

import { Loading, Error, Datagrid, TextField, DateField } from 'react-admin';

import {Grid, Card, CardContent, Typography} from "@mui/material";


/*
export const LastUpdatedDocumentsList = () => {
    const [page, setPage] = useState(1);
    const [perPage, setPerPage] = useState(10);
    const [sort, setSort] = useState({ field: "updated_at", order: "DESC" });
    const { data, total, loading, error } = useQueryWithStore({
      type: "getList",
      resource: "documents",
      payload: {
        pagination: { page, perPage },
        sort: {
          field: "created_at",
          order: "DESC",
        },
        filter: {},
      },
    });
  
    if (loading) {
      return <Loading />;
    }
    if (error) {
      return <Error error={error} />;
    }
  
    return (
      <Grid>
        <Card>
          <CardContent>
            <Typography variant="h5" color="textSecondary">
              Last updated documents
            </Typography>
              
            <Datagrid
              basePath="/documents"
              rowClick="show"
              isRowSelectable={false}
              data={keyBy(data, "id")}
              ids={data.map(({ id }) => id)}
              currentSort={sort}
            >
              <TextField source="name" />
              <DateField source="date" />
              <DateField source="created_at" />
            </Datagrid>
          </CardContent>
        </Card>
      </Grid>
    );
  };
  
  */