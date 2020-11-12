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
import Timeline from '@material-ui/lab/Timeline';
import TimelineItem from '@material-ui/lab/TimelineItem';
import TimelineSeparator from '@material-ui/lab/TimelineSeparator';
import TimelineConnector from '@material-ui/lab/TimelineConnector';
import TimelineContent from '@material-ui/lab/TimelineContent';
import TimelineDot from '@material-ui/lab/TimelineDot';
import TimelineOppositeContent from "@material-ui/lab/TimelineOppositeContent";

import {Card, Typography} from '@material-ui/core';


export const DocumentTimeline = (props) => {
    return (
        <Card>
            <Typography style={{padding: 16}} variant="h5" color="textSecondary">Documents timeline</Typography>
            <Timeline align="right">
                {props.stats.map( year =>
                    <TimelineItem>
                        <TimelineOppositeContent>
                            <Typography color="textSecondary">{year.num_documents} Documents</Typography>
                        </TimelineOppositeContent>
                        <TimelineSeparator>
                            <TimelineDot color={"primary"}/>
                            <TimelineConnector />
                        </TimelineSeparator>
                        <TimelineContent>{year.year}</TimelineContent>
                    </TimelineItem>
                )}
            </Timeline>
        </Card>
    );
}

