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

// Convert Numerical byte to string representation.
export const ByteToString = (byte: number | string): string => {
    const classes = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
    let val = Number(byte);
    if (val === NaN) {
        return byte.toString();
    }

    for (const key of classes) {
        if (val < 100) return `${Math.round(val * 100) / 100} ${key}`;
        if (val < 1024) return `${Math.round(val * 10) / 10} ${key}`;
        val = val / 1024;
    }
    return byte.toString();
}


export const LimitStringLength = (val: string, limit: number): string => {
    if (limit < 6) {
        return val;
    }
    if (val.length <= limit) {
        return val;
    }
    return val.substring(0, limit - 3).concat("...")
}