import React from "react";
import {
  Divider,
  Grid,
  Typography,
  ToggleButtonGroup,
  ToggleButton,
  Box,
} from "@mui/material";
import { TextField, useStore } from "react-admin";
import { MarkdownField } from "../markdown";

type ModeId = "text" | "rich-text";

export const ShowDocumentContent = () => {
  const [textMode, setTextMode] = useStore<ModeId>(
    "show-document.document-content.markdownMode",
    "text",
  );

  const handleChange = (_: React.MouseEvent<HTMLElement>, value: ModeId) => {
    setTextMode(value);
  };

  return (
    <Grid container alignItems={"center"} spacing={1}>
      <Grid item xs={7} sm={8} md={9}>
        <Typography variant="h5">Document content</Typography>
      </Grid>
      <Grid item xs={5} sm={4} md={3}>
        <ToggleButtonGroup value={textMode} exclusive onChange={handleChange}>
          <ToggleButton value={"text"} size={"small"} color={"primary"}>
            Text
          </ToggleButton>
          <ToggleButton value={"rich-text"} size={"small"} color={"primary"}>
            Rich text
          </ToggleButton>
        </ToggleButtonGroup>
      </Grid>
      <Grid item xs={12}>
        <Divider />
      </Grid>
      <Grid item>
        <Box
          sx={{
            maxHeight: "90vh",
            overflowY: "scroll",
          }}
        >
          {textMode === "text" ? (
            <TextField source="content" label="Raw parsed text content" />
          ) : (
            <MarkdownField source="content" label="Raw parsed text content" />
          )}
        </Box>
      </Grid>
    </Grid>
  );
};
