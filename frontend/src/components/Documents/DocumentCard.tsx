import {
  Card,
  CardActions,
  CardContent,
  CardHeader,
  ToggleButton,
  Typography,
  Chip,
} from "@mui/material";
import { EditButton, RaRecord, ShowButton } from "react-admin";
import { ThumbnailSmall } from "./Thumbnail";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import RadioButtonUncheckedIcon from "@mui/icons-material/RadioButtonUnchecked";
import * as React from "react";
import "./card.css";

const cardStyle = {
  width: 230,
  minHeight: 200,
  margin: "0.5em",
  display: "inline-block",
  verticalAlign: "top",
  borderRadius: 15,
  background: "#fafafc",
};

export const DocumentCard = (props: any) => {
  const { record } = props;
  const { selected, setSelected } = props;

  const isSelected = selected ? selected(record.id) : false;
  const select = () => (setSelected ? setSelected(record.id) : null);

  return (
    <Card key={record.id} style={{ ...cardStyle }}>
      <CardHeader
        title={<DocumentTitle record={record} />}
        sx={{ mt: 0.0, pb: 0 }}
      />
      <DocumentContent record={record} />
      <CardActions style={{ textAlign: "right", paddingTop: "0" }}>
        <ShowButton resource="documents" record={record} />
        <EditButton resource="documents" record={record} />
        <ToggleButton
          size="small"
          value={record.id}
          selected={isSelected}
          onChange={select}
          sx={{
            borderWidth: "0px",
            background: "primary",
            marginLeft: "auto",
          }}
        >
          {isSelected ? (
            <CheckCircleIcon color="primary" />
          ) : (
            <RadioButtonUncheckedIcon />
          )}
        </ToggleButton>
      </CardActions>
    </Card>
  );
};

const DocumentTitle = (props: { record: RaRecord }) => {
  const { record } = props;
  if (!record) return null;

  return (
    <Typography variant="subtitle2" sx={{ mt: 0.0, mb: 0 }}>
      <p className="document-title">{record.name}</p>
    </Typography>
  );
};

const DocumentContent = (props: { record: RaRecord }) => {
  const { record } = props;
  if (!record) return null;

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = (): colorTypes => {
    if (!record) {
      return "primary";
    }

    switch (record?.mimetype) {
      case "application/pdf":
        return "primary";
      case "text/plain":
        return "secondary";
      case "image/png":
      case "image/jpg":
      case "image/jpeg":
        return "success";
      default:
        return "warning";
    }
  };

  const getMimeTypeName = (): string => {
    if (!record) {
      return "";
    }

    switch (record?.mimetype) {
      case "application/pdf":
        return "PDF";
      case "text/plain":
        return "Text";
      case "image/png":
      case "image/jpg":
      case "image/jpeg":
        return "Image";
      default:
        return "";
    }
  };

  return (
    <CardContent style={{ position: "relative" }} sx={{ pt: 0.5 }}>
      <ThumbnailSmall url={record.preview_url} label="Img" />
      <Badge
        label={getDateString()}
        variant="outlined"
        color={"primary"}
        style={{ top: "4px", left: "16px", background: "white" }}
      />
      <Badge
        label={getMimeTypeName()}
        variant="filled"
        color={getMimetypeColor()}
        style={{ top: "4px", right: "16px" }}
      />
    </CardContent>
  );
};

type colorTypes =
  | "default"
  | "primary"
  | "secondary"
  | "error"
  | "info"
  | "success"
  | "warning";

const Badge = (props: {
  label: string;
  variant?: "filled" | "outlined";
  style?: object;
  color: colorTypes;
}) => {
  return (
    <Chip
      {...props}
      style={{ ...props.style, position: "absolute" }}
      size="small"
    />
  );
};
