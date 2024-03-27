/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import { defaultDarkTheme, defaultTheme } from "react-admin";
import { alpha } from "@mui/material";

export const lightTheme = {
  ...defaultTheme,
  palette: {
    logo: "#fff",
    favorite: { main: "#FFB300", contrastText: "#000" },
    primary: {
      main: "#673ab7",
    },
    secondary: {
      main: "#f50057",
      contrastText: "#fff",
    },
    background: {
      default: "#fcfcfe",
    },
    shape: {
      borderRadius: 10,
    },
    mode: "light" as const,
  },
  components: {
    MuiButtonBase: {
      defaultProps: {
        // disable ripple for perf reasons
        disableRipple: true,
      },
      styleOverrides: {
        root: {
          "&:hover:active::after": {
            // recreate a static ripple color
            // use the currentColor to make it work both for outlined and contained buttons
            // but to dim the background without dimming the text,
            // put another element on top with a limited opacity
            content: '""',
            display: "block",
            width: "100%",
            height: "100%",
            position: "absolute",
            top: 0,
            right: 0,
            backgroundColor: "currentColor",
            opacity: 0.3,
            borderRadius: "inherit",
          },
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        colorSecondary: {
          color: "#808080",
          backgroundColor: "#fff",
        },
      },
    },

    MuiToolBar: {
      styleOverrides: {
        colorSecondary: {
          color: "#808080",
          backgroundColor: "#fff",
        },
      },
    },

    MuiLinearProgress: {
      styleOverrides: {
        colorPrimary: {
          backgroundColor: "#f5f5f5",
        },
        barColorPrimary: {
          backgroundColor: "#d7d7d7",
        },
      },
    },
    MuiFilledInput: {
      styleOverrides: {
        root: {
          backgroundColor: "rgba(0, 0, 0, 0.04)",
          "&$disabled": {
            backgroundColor: "rgba(0, 0, 0, 0.04)",
          },
        },
      },
    },
    MuiSnackbarContent: {
      styleOverrides: {
        root: {
          border: "none",
        },
      },
    },
  },
};

const darkThemeBackgroundColor = "rgb(22,22,22)";

const darkThemePaperShadows = [
  alpha(darkThemeBackgroundColor, 0.2),
  alpha(darkThemeBackgroundColor, 0.1),
  alpha(darkThemeBackgroundColor, 0.05),
];

const elevationShadow = `${darkThemePaperShadows[0]} -2px 2px, ${darkThemePaperShadows[1]} -6px 6px,${darkThemePaperShadows[2]} -6px 6px`;
const shadow = { boxShadow: elevationShadow };

export const darkTheme = {
  ...defaultDarkTheme,
  palette: {
    ...defaultDarkTheme.palette,
    logo: "#673ab7",
    favorite: { main: "#FF8F00", contrastText: "#000" },
    primary: {
      main: "#a97aff",
    },
    background: {
      default: "#181818",
      secondary: "#673ab7",
    },
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        elevation1: shadow,
        elevation2: shadow,
        elevation3: shadow,
        elevation4: shadow,
        elevation5: shadow,
        elevation6: shadow,
        elevation7: shadow,
        elevation8: shadow,
        elevation9: shadow,
        elevation10: shadow,
        root: {
          backgroundClip: "padding-box",
        },
      },
    },
  },
};
