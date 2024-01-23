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

import * as dateFns from "date-fns";

// Convert Numerical byte to string representation.
export const ByteToString = (byte: number | string): string => {
  const classes = ["B", "KiB", "MiB", "GiB", "TiB", "PiB"];
  let val = Number(byte);
  if (Number.isNaN(val)) {
    return byte.toString();
  }

  for (const key of classes) {
    if (val < 100) return `${Math.round(val * 100) / 100} ${key}`;
    if (val < 1024) return `${Math.round(val * 10) / 10} ${key}`;
    val = val / 1024;
  }
  return byte.toString();
};

export const LimitStringLength = (val: string, limit: number): string => {
  if (limit < 6) {
    return val;
  }
  if (val.length <= limit) {
    return val;
  }
  return val.substring(0, limit - 3).concat("...");
};

export const EscapeWhitespace = (input: string): string => {
  if (input && input.includes(" ")) {
    return `"${input}"`;
  }
  return input;
};

// Prettify time since now().
//e.g. -> 'Just now', '30 minutes ago', '5 days ago', '7/7/2022'.
export const PrettifyRelativeTime = (time: number | string): string => {
  const now = Date.now();
  // @ts-ignore
  const secondsDiff =
    (now - (typeof time === "string" ? Date.parse(time) : time)) / 1000;
  if (secondsDiff < 60) {
    return "Just now";
  }

  const minutesDiff = Math.round(secondsDiff / 60);
  if (minutesDiff < 60) {
    return `${minutesDiff} minutes ago`;
  }

  const hoursDiff = Math.round(minutesDiff / 60);
  if (hoursDiff < 24) {
    return `${hoursDiff} hours ago`;
  }

  const daysDiff = Math.floor(hoursDiff / 24);
  if (daysDiff < 7) {
    return `${daysDiff} days ago`;
  }

  const date = new Date(time);
  return date.toLocaleDateString();
};

export const PrettifyAbsoluteTime = (time: number | string): string => {
  const d = typeof time === "string" ? Date.parse(time) : time;
  return dateFns.formatRelative(d, new Date());
};

export const PrettifyTimeInterval = (
  startTime: number | string,
  stopTime: number | string,
): string => {
  const start =
    typeof startTime === "string" ? Date.parse(startTime) : startTime;
  const stop = typeof stopTime === "string" ? Date.parse(stopTime) : stopTime;

  const duration = dateFns.intervalToDuration({ start, end: stop });
  const formatted = dateFns.formatDuration(duration);
  return formatted;
};
