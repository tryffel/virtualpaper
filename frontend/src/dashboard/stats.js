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
import {Card, Typography, Box} from '@material-ui/core';
import {makeStyles} from "@material-ui/core/styles";


const useStyles = makeStyles(() => ({

    card: {
        minHeight: 30,
        display: 'flex',
        flexDirection: 'column',
        flex: '1',
        '& a': {
            textDecoration: 'none',
            color: 'inherit',
        },
    },
    main: (props) => ({
        overflow: 'inherit',
        padding: 16,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
    }),
    title: {},
}));


export const SingleStat = (props) => {
    const classes = useStyles(props);

    return (
        <Card className={classes.card}>
            <div className={classes.main}>
                <Box textAlign="right" >
                    <Typography
                        variant="h5"
                        color="textSecondary"
                    >
                        {props.title}
                    </Typography>
                    <Typography variant="h5" component="h2">
                        {props.count|| ' '}
                    </Typography>
                </Box>
            </div>
        </Card>
    )
}


export const Stats = (data) => {

    return (
        <Card>
            <SingleStat title={"Total documents"} count={data.num_documents}/>
            <SingleStat title={"Metadata keys"} count={data.num_metadata_keys}/>
            <SingleStat title={"Metadata values"} count={data.num_metadata_values}/>
        </Card>
    )
}