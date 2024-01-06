import { Button, LabeledClasses, useRecordContext } from "react-admin";
import get from "lodash/get";
import { Grid, Tooltip, Typography } from "@mui/material";
import List from "@mui/material/List";
import { IconByName, iconExists } from "../../../components/icons";
import LabelIcon from "@mui/icons-material/Label";
import ListItem from "@mui/material/ListItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import React from "react";
import { EscapeWhitespace } from "../../../components/util";
import { Link } from "react-router-dom";

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
    <>
      <Typography className={LabeledClasses.label}>Metadata</Typography>
      <List dense>
        {array.map((item) => {
          return <MetadataValue metadata={item} />;
        })}
      </List>
    </>
  );
};

const MetadataValue = ({ metadata }: { metadata: Metadata }) => {
  const style = JSON.parse(metadata.style);
  const color = get(style, "color") ?? "inherit";
  const icon = iconExists(metadata.icon) ? (
    <IconByName name={metadata.icon} color={color} />
  ) : (
    <LabelIcon color={color} />
  );

  const linkDocsLabel = `Show documents`;

  let to: any = {
    pathname: "/documents",
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
                />
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
