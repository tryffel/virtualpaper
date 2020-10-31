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


import {Card as MuiCard, CardContent, withStyles} from "@material-ui/core";
import {FilterList, FilterListItem, FilterLiveSearch} from "react-admin";
import AccessTimeIcon from "@material-ui/icons/AccessTime";
import {endOfYesterday, startOfMonth, startOfWeek, startOfYear, subYears} from "date-fns";
import * as React from "react";


const Card = withStyles(theme => ({
    root: {
        [theme.breakpoints.up('sm')]: {
            order: -1,
            width: '15em',
            marginRight: '1em',
        },
        [theme.breakpoints.down('sm')]: {
            display: 'none',
        },
    },
}))(MuiCard);



export const FilterSidebar = () => (
    <Card style={{width: '20em', minWidth: '15em'}}>
        <CardContent>
            <FilterLiveSearch source="q" />
            <LastVisitedFilter/>
        </CardContent>
    </Card>
);

const LastVisitedFilter = () => (
    <FilterList label="Last visited" icon={<AccessTimeIcon />} >
        <FilterListItem
            label="Today"
            value={{
                after: endOfYesterday().getTime(),
                before: undefined,
            }}
        />
        <FilterListItem
            label="This week"
            value={{
                after: startOfWeek(new Date()).getTime(),
                before: undefined,
            }}
        />
        <FilterListItem
            label="This month"
            value={{
                after: startOfMonth(new Date()).getTime(),
                before: undefined,
            }}
        />
        <FilterListItem
            label="This year"
            value={{
                after: startOfYear(new Date()).getTime(),
                before: undefined,
            }}
        />
        <FilterListItem
            label="Last year"
            value={{
                after: subYears(startOfYear(new Date()),1).getTime(),
                before: startOfYear(new Date()).getTime(),
            }}
        />
        <FilterListItem
            label="Earlier"
            value={{
                after: undefined,
                before: subYears(startOfYear(new Date()),2).getTime(),
            }}
        />
    </FilterList>
);
