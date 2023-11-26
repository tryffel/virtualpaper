import React from "react";
import { TextInput } from "react-admin";

export const StopWordsInput = () => {
  const parse = (value: string) => {
    if (!value) {
      return "[]";
    }
    return value.split("\n");
  };

  const format = (value: string | Array<string>) => {
    if (typeof value === "string") {
      // @ts-ignore
      const data: Array<string> = JSON.parse(value);
      return data.join("\n");
    } else {
      // @ts-ignore
      return value.join("\n");
    }
  };

  return (
    <TextInput
      multiline
      fullWidth
      source="stop_words"
      label={""}
      parse={parse}
      format={format}
    />
  );
};

export const SynonymsInput = () => {
  const parse = (value: string) => {
    if (!value) {
      return "[]";
    }
    const rows = value.split("\n");
    const items = rows.map((row) => row.split(","));
    return items;
  };

  const format = (value: string | Array<Array<string>>) => {
    if (!value) {
      return "";
    }
    if (typeof value === "string") {
      // @ts-ignore
      const data: Array<Array<string>> = JSON.parse(value);
      return data.map((row) => row.join(",")).join("\n");
    } else {
      // @ts-ignore
      return value.map((row) => row.join(",")).join("\n");
    }
  };
  return (
    <TextInput
      multiline
      fullWidth
      source="synonyms"
      label={""}
      parse={parse}
      format={format}
    />
  );
};
