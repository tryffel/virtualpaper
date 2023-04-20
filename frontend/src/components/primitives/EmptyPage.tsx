import { Grid, Typography } from "@mui/material";
import { CreateButton } from "react-admin";
import * as React from "react";

import InboxIcon from "@mui/icons-material/Inbox";

export interface EmptyPageProps {
  title: string;
  subTitle?: string;
  noCreateButton?: boolean;
  createButtonLabel?: string;
  icon?: React.ReactElement;
  resource?: string;
}

export const EmptyResourceIconSx = {
  width: "9em",
  height: "9em",
  opacity: 0.5,
};

export const EmptyResourcePage = (props: EmptyPageProps) => {
  const finalProps = {
    ...props,
    subTitle: props.subTitle ?? "Do you want to create one?",
    createButtonLabel: props.createButtonLabel ?? "Create",
    icon: props.icon ? (
      React.cloneElement(props.icon, { sx: EmptyResourceIconSx })
    ) : (
      <InboxIcon sx={EmptyResourceIconSx} />
    ),
  };

  return (
    <Grid
      container
      direction="column"
      justifyContent="center"
      alignItems="center"
      sx={{ mt: "20px" }}
    >
      {finalProps.icon && finalProps.icon}

      <Typography
        variant="h4"
        paragraph
        style={{ opacity: 0.5, marginTop: "10px" }}
      >
        {finalProps.title}
      </Typography>
      {finalProps.subTitle && (
        <Typography
          variant="body1"
          paragraph
          sx={{ opacity: 0.5, marginTop: "10px" }}
        >
          {finalProps.subTitle}
        </Typography>
      )}
      {!finalProps.noCreateButton && (
        <CreateButton
          resource={props.resource}
          label={finalProps.createButtonLabel}
          variant={"contained"}
          sx={{ marginTop: "10px" }}
        />
      )}
    </Grid>
  );
};
