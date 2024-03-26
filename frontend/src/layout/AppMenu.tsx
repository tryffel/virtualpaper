import { Menu, ThemesContext, ThemeProvider } from "react-admin";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import { AppSidebar } from "./SideBar.tsx";
import { Typography } from "@mui/material";
import { darkTheme } from "../theme.ts";

const listItemStyle = {
  pl: 0,
  pt: 0,
  pb: 0,
};

export const AppMenu = () => {
  return (
    <ThemesContext.Provider value={{ lightTheme: darkTheme, darkTheme }}>
      <ThemeProvider>
        <AppSidebar>
          <Menu className={"RaAppMenu"}>
            <ListItem sx={listItemStyle}>
              <List>
                <ListItem>
                  <ListItemText>Documents</ListItemText>
                </ListItem>
                <Menu.DashboardItem
                  primaryText={<Typography>Dashboard</Typography>}
                />
                <Menu.ResourceItem name={"documents"} />
                <Menu.ResourceItem name={"documents/deleted"} />
              </List>
            </ListItem>
            <ListItem sx={listItemStyle}>
              <List>
                <ListItem>
                  <ListItemText>Metadata</ListItemText>
                </ListItem>
                <Menu.ResourceItem name={"metadata/keys"} />
              </List>
            </ListItem>
            <ListItem sx={listItemStyle}>
              <List>
                <ListItem>Processing</ListItem>
                <Menu.ResourceItem name={"processing/rules"} />
              </List>
            </ListItem>
          </Menu>
        </AppSidebar>
      </ThemeProvider>
    </ThemesContext.Provider>
  );
};
