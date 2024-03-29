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

import { useRecordContext, useTheme } from "react-admin";
import { Box } from "@mui/material";
import MDEditor from "@uiw/react-md-editor/nohighlight";

import get from "lodash/get";

export type MarkdownProps = {
  source: string;
  label?: string;
};

export const MarkdownField = (props: MarkdownProps) => {
  const record = useRecordContext(props);
  const theme = useTheme();

  return (
    <Box data-color-mode={theme} sx={{ p: 1 }}>
      <MDEditor.Markdown source={get(record, props.source)} />
    </Box>
  );
};

export default MarkdownField;
