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

import React from "react";
import ReactMarkdown from "react-markdown";
import { Labeled, useInput, useRecordContext } from "react-admin";
import { useController } from "react-hook-form";
import * as Showdown from "showdown";
//import "react-mde/lib/styles/css/react-mde-all.css";

//import ReactMde from "react-mde";

export const MarkdownField = (props: any) => {
  const record = useRecordContext(props);

  return <ReactMarkdown>{record[props.source]}</ReactMarkdown>;
};

// Adapted from github.com/maluramichael/ra-input-markdown, adding support for RA4. Originally licensed under MIT.
export const MarkdownInput = (props: any) => {
  const { onChange, onBlur, ...rest } = props;
  const {
    field,
    fieldState: { isTouched, invalid, error },
    formState: { isSubmitted },
  } = useInput({onChange, onBlur, ...props});

  const [text, setText] = React.useState("- original texten");

  const converter = new Showdown.Converter({
    simplifiedautolink: true,
    strikethrough: true,
    tasklists: true
  });
  
  const [selectedTab, setSelectedTab] = React.useState<"write" | "preview">("write");
  return (
    <Labeled label={props.label}>
      <></>
      {/*
      <ReactMde
        value={field.value}
        onChange={field.onChange}
        selectedTab={selectedTab}
        onTabChange={setSelectedTab}
        generateMarkdownPreview={(md) => Promise.resolve(converter.makeHtml(md))}
      />
      */}
    </Labeled>

  );
};
