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

import ReactMarkdown from "react-markdown";
import {Labeled, useInput, useRecordContext, useTheme} from "react-admin";
import {Box} from "@mui/material";
import MDEditor from '@uiw/react-md-editor';

export const MarkdownField = (props: any) => {
  const record = useRecordContext(props);
  return <ReactMarkdown>{record[props.source]}</ReactMarkdown>;
};

export const MarkdownInput = (props: any) => {
  const theme = useTheme()
  const { onChange, onBlur} = props;
  const { field} = useInput({onChange, onBlur, ...props});
  return (
    <Labeled label={props.label} fullWidth>
      <Box data-color-mode={theme}>
        <MDEditor
            value={field.value}
            onChange={field.onChange}
            preview="edit"
          />
      </Box>
    </Labeled>
  );
};
