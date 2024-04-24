import { Datagrid, List, TextField } from "react-admin";

export const PropertyList = () => {
  return (
    <List>
      <Datagrid rowClick={"edit"}>
        <TextField source={"name"} />
      </Datagrid>
    </List>
  );
};
