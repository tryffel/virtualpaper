import {
  Card,
  CardActions,
  CardActionArea,
  CardContent,
  CardHeader,
  ToggleButton,
  Typography,
  Chip,
} from "@mui/material";
import RestoreIcon from "@mui/icons-material/Restore";
import {
  Button,
  Confirm,
  EditButton,
  RaRecord,
  useDelete,
  useNotify,
  useRefresh,
  useTheme,
  useUpdate,
} from "react-admin";
import { ThumbnailSmall } from "./Thumbnail";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import RadioButtonUncheckedIcon from "@mui/icons-material/RadioButtonUnchecked";
import "./document.css";
import get from "lodash/get";
import { useState } from "react";
import DeleteIcon from "@mui/icons-material/Delete";
import ShareIcon from "@mui/icons-material/Share";

const cardStyle = {
  width: 230,
  minHeight: 200,
  margin: "0.5em",
  display: "inline-block",
  verticalAlign: "top",
  borderRadius: 15,
  background: "secondary",
};

export interface DocumentCardProps {
  record: RaRecord;
  selected?: (id: string) => boolean;
  setSelected?: (id: string) => void;
}

export const DocumentCard = (props: DocumentCardProps) => {
  const [theme] = useTheme();

  const { record } = props;
  const { selected, setSelected } = props;
  const isDeleted = get(record, "deleted_at") !== null;

  const isSelected = selected ? selected(String(record.id)) : false;
  const select = () => (setSelected ? setSelected(String(record.id)) : null);

  return (
    <Card
      key={record.id}
      style={{ ...cardStyle }}
      sx={{
        backgroundColor: theme === "dark" ? "#272727" : "#FAFAFA",
        border:
          theme === "dark"
            ? "1px solid rgba(255, 255, 255, 0.12)"
            : "1px solid rgba(0, 0, 0, 0.12)",
      }}
    >
      <CardActionArea href={`/#/documents/${record.id}/show`}>
        <CardHeader
          title={<DocumentTitle record={record} />}
          sx={{ mt: 0, pb: 0, pt: 0, height: 85 }}
        />
        <DocumentContent record={record} />
      </CardActionArea>
      <CardActions style={{ textAlign: "right", paddingTop: "0" }}>
        {!isDeleted ? (
          <>
            <EditButton resource="documents" record={record} />
            {setSelected && (
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
            )}
          </>
        ) : (
          <>
            <RestoreDocumentButton record={record} />
            <ConfirmDeleteDocumentButton record={record} />
          </>
        )}
      </CardActions>
    </Card>
  );
};

const DocumentTitle = (props: { record: RaRecord }) => {
  const { record } = props;
  if (!record) return null;

  return (
    <Typography variant="subtitle2" sx={{ mt: 0.0, mb: 0, pt: 0 }}>
      <div className="document-title">
        <p
          dangerouslySetInnerHTML={{
            __html: record.name,
          }}
        />
      </div>
    </Typography>
  );
};

const DocumentContent = (props: { record: RaRecord }) => {
  const [theme] = useTheme();
  const { record } = props;
  if (!record) return null;

  const getDateString = (): string => {
    if (!record) {
      return "";
    }
    const date = new Date(record.date);
    return date.toLocaleDateString();
  };

  const getMimetypeColor = (): colorTypes => mimetypeToColor(record?.mimetype);
  const getMimeTypeName = (): string => mimetypeToText(record?.mimetype);

  const shared = get(record, "shares") > 0;

  return (
    <CardContent
      style={{ position: "relative" }}
      sx={{
        pt: 0.5,
        backgroundColor: theme === "dark" ? "#D9D9D9" : "#f8f8f8",
      }}
    >
      <ThumbnailSmall url={record.preview_url} label="Img" />
      <Badge
        label={getDateString()}
        variant="outlined"
        color={"primary"}
        style={{
          top: "4px",
          left: "16px",
          background: "white",
          position: "absolute",
        }}
      />
      <Badge
        label={getMimeTypeName()}
        variant="filled"
        color={getMimetypeColor()}
        style={{ top: "4px", right: "16px", position: "absolute" }}
      />
      {shared && (
        <Chip
          icon={<ShareIcon />}
          label={"Shared"}
          variant="filled"
          color={"warning"}
          style={{
            top: "35px",
            left: "16px",
            position: "absolute",
            height: "24px",
          }}
        />
      )}
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

export const Badge = (props: {
  label: string;
  variant?: "filled" | "outlined";
  style?: object;
  color: colorTypes;
  sx?: object;
}) => {
  return (
    <Chip {...props} style={{ ...props.style }} size="small" sx={props.sx} />
  );
};

export const mimetypeToText = (mimetype: string) => {
  switch (mimetype) {
    case "application/pdf":
      return "PDF";
    case "text/plain":
      return "Text";
    case "image/png":
    case "image/jpg":
    case "image/jpeg":
      return "Image";
    default:
      return "Misc";
  }
};

export const mimetypeToColor = (mimetype: string) => {
  switch (mimetype) {
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

const RestoreDocumentButton = (props: { record: RaRecord }) => {
  const { record } = props;

  const refresh = useRefresh();
  const notify = useNotify();
  const [confirmOpen, setConfirmOpen] = useState(false);

  const [update, { isLoading, error }] = useUpdate("documents/deleted", {
    id: get(record, "id"),
    data: {},
    previousData: record,
    meta: { action: "restore" },
  });

  const handleClick = () => setConfirmOpen(true);
  const handleDialogClose = () => setConfirmOpen(false);
  const handleConfirm = async () => {
    await update();
    setTimeout(() => refresh(), 1000);
    if (error) {
      // @ts-ignore
      notify(error.toString(), { type: "error" });
    } else {
      setConfirmOpen(false);
      notify("Document restored");
    }
  };

  return (
    <>
      <Button label={"Restore"} color={"primary"} onClick={handleClick}>
        <RestoreIcon />
      </Button>
      <Confirm
        isOpen={confirmOpen}
        loading={isLoading}
        title="Restore document"
        content="Are you sure you want to restore the document?"
        onConfirm={handleConfirm}
        onClose={handleDialogClose}
      />
    </>
  );
};

const ConfirmDeleteDocumentButton = (props: { record: RaRecord }) => {
  const { record } = props;

  const refresh = useRefresh();
  const notify = useNotify();
  const [confirmOpen, setConfirmOpen] = useState(false);

  const [update, { isLoading, error }] = useDelete("documents/deleted", {
    id: get(record, "id"),
    previousData: record,
  });

  const handleClick = () => setConfirmOpen(true);
  const handleDialogClose = () => setConfirmOpen(false);
  const handleConfirm = async () => {
    await update();
    setTimeout(() => refresh(), 1000);
    if (error) {
      // @ts-ignore
      notify(error.toString(), { type: "error" });
    } else {
      setConfirmOpen(false);
      notify("Document deleted");
    }
  };

  return (
    <>
      <Button label={"Delete"} color={"secondary"} onClick={handleClick}>
        <DeleteIcon />
      </Button>
      <Confirm
        isOpen={confirmOpen}
        loading={isLoading}
        title="Restore document"
        content="Are you sure you want to permanently delete the document?"
        onConfirm={handleConfirm}
        onClose={handleDialogClose}
      />
    </>
  );
};
