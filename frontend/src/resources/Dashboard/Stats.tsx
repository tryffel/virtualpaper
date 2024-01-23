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

import * as React from "react";
import {Card, Typography, CardContent, Grid} from '@mui/material';


export interface StatProps {
    title: string;
    count: string;
}


export const SingleStat = (props: StatProps) => {

    return (
        <Grid item xs={4} sm={4}>
            <Card>
                <CardContent>
                    <Typography
                        variant="h6"
                        color="textSecondary"
                    >
                        {props.title}
                    </Typography>
                    <Typography variant="h6" component="h4">
                        {props.count|| ' '}
                    </Typography>
                </CardContent>
            </Card>
        </Grid>
    )
}


export const Stats = (data: any) => {

    return (
        <Grid container spacing={1}>
            <SingleStat title={"Total documents"} count={data.num_documents}/>
            <SingleStat title={"Metadata keys"} count={data.num_metadata_keys}/>
            <SingleStat title={"Metadata values"} count={data.num_metadata_values}/>
        </Grid>
    )
}