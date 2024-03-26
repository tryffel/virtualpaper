import * as React from "react";
import { Drawer } from "@mui/material";
import { SidebarClasses, useLocale, useSidebarState } from "react-admin";

// Image thanks to Christa Dodoo: https://unsplash.com/@krystagrusseck?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash /
// https://unsplash.com/photos/pile-of-papers-MldQeWmF2_g?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash
//
import image from "./papers.jpg";

export const AppSidebar = ({ children }: React.PropsWithChildren) => {
  const [open, setOpen] = useSidebarState();
  useLocale(); // force redraw on locale change

  const toggleSidebar = () => setOpen(!open);

  return (
    <Drawer
      variant="temporary"
      open={open}
      onClose={toggleSidebar}
      classes={{ ...SidebarClasses }}
      sx={{
        display: "block",
        "& .MuiDrawer-paper": {
          boxSizing: "border-box",
          backgroundImage: `linear-gradient( rgba(0, 0, 0, 0.8), rgba(0, 0, 0, 0.8) ),url(${image})`,
          backdropFilter: "blur(100px)",
          position: "absolute",
          backgroundSize: "cover",
          backgroundPosition: "center center",
        },
      }}
    >
      {children}
    </Drawer>
  );
};
