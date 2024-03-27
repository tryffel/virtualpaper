import React, { Suspense } from "react";
import { Button, Labeled, useRecordContext } from "react-admin";
import get from "lodash/get";
import { Grid, Tooltip, Typography } from "@mui/material";
import List from "@mui/material/List";
const IconByName = React.lazy(() => import("../../../components/icons.tsx"));
import ListItem from "@mui/material/ListItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import { EscapeWhitespace } from "../../../components/util";
import { Link } from "react-router-dom";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";

type Metadata = {
  id: number;
  key: string;
  value: string;
  icon: string;
  style: string;
};

export const MetadataList = () => {
  const record = useRecordContext();
  if (!record) {
    return null;
  }

  const array: Metadata[] = get(record, "metadata");
  if (!array) {
    return null;
  }

  return (
    <Labeled label={"Metadata"}>
      <List dense>
        {array.map((item) => {
          return <MetadataValue metadata={item} />;
        })}
      </List>
    </Labeled>
  );
};

const MetadataValue = ({ metadata }: { metadata: Metadata }) => {
  if (!metadata || !metadata.style) {
    return null;
  }
  const style = JSON.parse(metadata.style);
  const color = get(style, "color") ?? "inherit";

  const icon = (
    <Suspense fallback={null}>
      <IconByName name={metadata.icon} color={color} />
    </Suspense>
  );

  const linkDocsLabel = `Show documents`;

  let to = {
    pathname: "/documents",
    search: "",
  };

  if (metadata) {
    to = {
      pathname: "/documents",
      search: `filter=${JSON.stringify({
        q:
          EscapeWhitespace(get(metadata, "key")) +
          ":" +
          EscapeWhitespace(get(metadata, "value")),
      })}`,
    };
  }
  return (
    <ListItem key={metadata.id}>
      <ListItemIcon>{icon}</ListItemIcon>
      <ListItemText
        primary={
          <Tooltip
            title={
              <Grid
                container
                spacing={2}
                flexDirection={"column"}
                sx={{ backgroundColor: "background" }}
              >
                <Grid item>
                  <Typography variant={"body1"}>
                    Show all documents with this value:
                  </Typography>
                </Grid>
                <Button
                  variant={"text"}
                  label={linkDocsLabel}
                  color={"primary"}
                  component={Link}
                  to={to}
                >
                  <OpenInNewIcon />
                </Button>
              </Grid>
            }
          >
            <Typography variant={"body2"}>
              {metadata.key}: {metadata.value}
            </Typography>
          </Tooltip>
        }
      />
    </ListItem>
  );
};
