import { Labeled, useRecordContext } from "react-admin";
import get from "lodash/get";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import { Typography } from "@mui/material";

export const DocumentPropertyList = () => {
  const record = useRecordContext();
  const properties = get(record, "properties");
  if (!properties) {
    return null;
  }

  return (
    <Labeled label={"Properties"}>
      <List>
        {properties.map((property: object) => (
          <ListItem key={get(property, "id")}>
            <ListItemText>
              <Typography variant={"body2"}>
                {get(property, "property_name")}: {get(property, "value")}
              </Typography>
            </ListItemText>
          </ListItem>
        ))}
      </List>
    </Labeled>
  );
};
