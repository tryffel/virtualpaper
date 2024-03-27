import { TextField, useRecordContext } from "react-admin";
import get from "lodash/get";
import {
  Badge,
  mimetypeToColor,
  mimetypeToText,
} from "@components/document/card";
import { Box, Grid } from "@mui/material";
import { DocumentIdField } from "@components/document/fields/DocumentId.tsx";
import { languages } from "@/languages.ts";

function getLanguageLabel(code: string): string {
  const lang = languages[code as keyof typeof languages];
  return lang ? (lang as string) : code;
}

export const DocumentTopRow = () => {
  const record = useRecordContext();
  const isDeleted = get(record, "deleted_at") !== null;
  const hasLang = get(record, "lang") !== "";

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = () => mimetypeToColor(record?.mimetype);
  const getMimeTypeName = (): string => mimetypeToText(record?.mimetype);
  const getLang = () => getLanguageLabel(record?.lang);
  const isFavorite = record && get(record, "favorite");


    return (
    <Grid container justifyContent={"align-center"}>
      <Grid item>
        <TextField source="name" label="" style={{ fontSize: "2em" }} />
      </Grid>

      <Grid item xs={12}>
        <Badge
          label={getDateString()}
          variant="outlined"
          color={"primary"}
          style={{
            top: "4px",
            left: "16px",
            background: "white",
          }}
          sx={{ m: 0.5 }}
        />
        {hasLang && (
          <Badge
            label={getLang()}
            variant="outlined"
            color={"primary"}
            sx={{ m: 0.5 }}
          />
        )}
        <Badge
          label={getMimeTypeName()}
          variant="filled"
          color={getMimetypeColor()}
          style={{ top: "4px", right: "16px" }}
          sx={{ m: 1 }}
        />
        {isFavorite && (
          <Badge
            label={"Favorite"}
            variant="filled"
            // @ts-ignore
            color={"favorite"}
            style={{ top: "4px", right: "16px" }}
            sx={{ m: 1 }}
          />
        )}
        {isDeleted && (
          <Badge
            label={"Document is deleted"}
            variant="filled"
            color={"error"}
            style={{ top: "4px", right: "16px" }}
            sx={{ m: 1 }}
          />
        )}
      </Grid>
      <Grid item xs={12}>
        <DocumentIdField />
      </Grid>
    </Grid>
  );
};
export const DocumentBasicInfo = () => {
  const record = useRecordContext();
  const isDeleted = get(record, "deleted_at") !== null;
  const hasLang = get(record, "lang") !== "";

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = () => mimetypeToColor(record?.mimetype);
  const getMimeTypeName = (): string => mimetypeToText(record?.mimetype);
  const getLang = () => getLanguageLabel(record?.lang);
  const isFavorite = record && get(record, "favorite");

  return (
    <Box>
      <Badge
        label={getDateString()}
        variant="outlined"
        color={"primary"}
        style={{
          top: "4px",
          left: "16px",
          background: "white",
          marginLeft: 0,
        }}
        sx={{ m: 0.5 }}
      />
      {hasLang && (
        <Badge
          label={getLang()}
          variant="outlined"
          color={"primary"}
          sx={{ m: 0.5 }}
        />
      )}
      <Badge
        label={getMimeTypeName()}
        variant="filled"
        color={getMimetypeColor()}
        style={{ top: "4px", right: "16px" }}
        sx={{ m: 1 }}
      />
      {isFavorite && (
        <Badge
          label={"Favorite"}
          variant="filled"
          // @ts-ignore
          color={"favorite"}
          style={{ top: "4px", right: "16px" }}
          sx={{ m: 1 }}
        />
      )}
      {isDeleted && (
        <Badge
          label={"Document is deleted"}
          variant="filled"
          color={"error"}
          style={{ top: "4px", right: "16px" }}
          sx={{ m: 1 }}
        />
      )}
    </Box>
  );
};
export const DocumentTitle = () => {
  return (
    <TextField
      source="name"
      label=""
      style={{ fontSize: "2em", paddingBottom: "0.8em" }}
    />
  );
};
