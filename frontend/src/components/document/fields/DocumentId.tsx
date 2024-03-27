import { useNotify, useRecordContext } from "react-admin";
import { Box, IconButton, Typography } from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";

export const DocumentIdField = () => {
  const record = useRecordContext();
  const notify = useNotify();
  const id = record ? record.id : "";

  const handleCopy = () => {
    if (id) {
      navigator.clipboard.writeText(id as string);
      notify("Id copied to clipboard", {
        type: "info",
        autoHideDuration: 3000,
      });
    }
  };

  return (
    <Box>
      <Typography
        component={"span"}
        style={{ userSelect: "none" }}
        variant={"body2"}
        color={"text.secondary"}
      >
        Id:
      </Typography>
      <Typography
        component={"span"}
        variant={"body2"}
        color={"text.secondary"}
        ml={1}
      >
        {id}
      </Typography>
      <IconButton sx={{ ml: 0.5 }} size={"small"}>
        <ContentCopyIcon
          fontSize={"small"}
          style={{ height: "18px", width: "18px" }}
          onClick={handleCopy}
          color={"primary"}
        />
      </IconButton>
    </Box>
  );
};
