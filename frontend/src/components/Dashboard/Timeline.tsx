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

import { Timeline } from '@mui/lab';
import TimelineItem from '@mui/lab/TimelineItem';
import TimelineSeparator from '@mui/lab/TimelineSeparator';
import TimelineConnector from '@mui/lab/TimelineConnector';
import TimelineContent from '@mui/lab/TimelineContent';
import TimelineDot from '@mui/lab/TimelineDot';
import TimelineOppositeContent from "@mui/lab/TimelineOppositeContent";
import {Link} from 'react-router-dom';
import {startOfYear, endOfYear} from "date-fns";

import {Card, Typography} from '@mui/material';

import {Loading} from "react-admin";


export const DocumentTimeline = (stats: any) => {
    const getDocumentsLink = (year: number) => {
        const d = new Date(year, 1, 1);
        const after = startOfYear(d).getTime();
        const before = endOfYear(d).getTime();
        return {
            pathname: "/documents",
            search: `filter=${JSON.stringify({ after: after, before: before })}`,
        }
    }
    
    console.log(stats);
    
    if (!stats) {return <Loading/>}

    return (
        <Card>
                   <Typography  variant="h5" color="textSecondary">Documents timeline</Typography>
            <Timeline >
            // @ts-ignore
                {stats.map( (year: { year: string | number | boolean | React.ReactElement<any, string | React.JSXElementConstructor<any>> | React.ReactFragment | null | undefined; num_documents: string | number | boolean | React.ReactElement<any, string | React.JSXElementConstructor<any>> | React.ReactFragment | null | undefined; }) =>
                    <TimelineItem>
                        <TimelineOppositeContent>
                                                    <Typography component={Link} 
// @ts-ignore

                            to={getDocumentsLink(year.year)}>
                            // @ts-ignore
                                    {year.num_documents} {year.num_documents === 1? "Document": "Documents"}
                            </Typography>
                        </TimelineOppositeContent>
                        <TimelineSeparator>
                            <TimelineDot color={"primary"}/>
                            <TimelineConnector />
                        </TimelineSeparator>
                        // @ts-ignore
                        <TimelineContent>{year.year}</TimelineContent>
                    </TimelineItem>
                )}
            </Timeline>
        </Card>
        
    );
}

