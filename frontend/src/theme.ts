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

export const lightTheme = {
  ...defaultTheme,
  palette: {
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
    mode: "light" as "light",
  },
  components: {
    /*
    RaMenuItemLink: {
      styleOverrides: {
        root: {
          borderLeft: "3px solid #fff",
          "&.RaMenuItemLink-active": {
            borderLeft: "3px solid #4f3cc9",
          },
        },
      },
    },

     */
    MuiPaper: {
      styleOverrides: {
        elevation1: {
          boxShadow: "none",
        },
        root: {
          border: "1px solid #e0e0e3",
          backgroundClip: "padding-box",
        },
      },
    },

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

export const darkTheme = {
  ...defaultDarkTheme,
  palette: {
    ...defaultDarkTheme.palette,
    primary: {
      main: "#673ab7",
    },
    background: {
      default: "#313131",
      secondary: "#673ab7",
    },
  },
};
