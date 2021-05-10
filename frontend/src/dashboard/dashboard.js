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


import * as React from "react";
import Card from '@material-ui/core/Card';
import {Box, Grid, useMediaQuery} from '@material-ui/core';
import {Error, Loading, useQueryWithStore} from 'react-admin';

import { Stats } from "./stats";
import { DocumentTimeline } from "./timeline";
import {LatestDocumentsList} from "../documents/list";


export default (props) => {
    const {data, loading, error } = useQueryWithStore({
        type: 'getOne',
        resource: 'documents/stats',
        payload: { target:"documents/stats", sort:"id", order:"asc"},
    });

    if (loading) return <Loading />;
    if (error) return <Error error={error}/>;

    let direction = "row";
    return (

            <Grid container spacing={1} direction={direction} flexGrow={1} alignItems="stretch">
                <Grid item xl={6} lg={6} sm={12} md={10} xs={12}>
                    <LatestDocumentsList{...props}/>
                </Grid>
                <Grid item xl={2} lg={5} xs={12} sm={10} md={8}>
                    <DocumentTimeline stats={data.yearly_stats}/>
                </Grid>
                <Grid item xs={12} sm={10} md={8} lg={3}>
                    <Stats {...data}/>
                </Grid>

        </Grid>
    );
}
