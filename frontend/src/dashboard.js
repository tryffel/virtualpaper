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
import { Box } from '@material-ui/core';
import CardContent from '@material-ui/core/CardContent';
import {Error, Loading, SelectInput, Title, useQueryWithStore} from 'react-admin';


export default () => {
    const {data, loading, error } = useQueryWithStore({
        type: 'getOne',
        resource: 'documents/stats',
        payload: { target:"documents/stats", sort:"id", order:"asc"},
    });

    if (loading) return <Loading />;
    if (error) return <Error error={error}/>;

    return (
        <Card>
            <Box display="flex">
                <Box flex="1">
                    <Card>
                        <CardContent>Total documents: {data.num_documents}</CardContent>
                    </Card>

                    <Card>
                        <CardContent>Total metadata keys: {data.num_metadata_keys}</CardContent>
                    </Card>

                    <Card>
                        <CardContent>Total metadata values: {data.num_metadata_values}</CardContent>
                    </Card>
                </Box>
                <Box flex="2">
                    <Card>
                        <CardContent>Document count by year</CardContent>
                    </Card>
                    {data.yearly_stats.map (year =>
                        <Box>
                            <Card>
                                <CardContent>{year.year}: {year.num_documents}</CardContent>
                            </Card>
                        </Box>)
                    }
                </Box>
            </Box>
        </Card>
    );
}
