import * as React from "react";
import "./App.css";

import TagIcon from "@mui/icons-material/Tag";
import ConstructionIcon from "@mui/icons-material/Construction";
import ArticleIcon from "@mui/icons-material/Article";

import { Route } from "react-router-dom";

import {
  Admin,
  ListGuesser,
  Resource,
  useTheme,
  CustomRoutes,
} from "react-admin";
import { dataProvider } from "./api/dataProvider";
import authProvider from "./api/authProvider";
import documents from "./components/Documents";

import { lightTheme, darkTheme } from "./theme";
import Layout from "./layout/Layout";
import MetadataKeys from "./components/MetadataKeys";
import Rules from "./components/Rules";

import { ProfileEdit } from "./components/Preferences";
import AdminView from "./components/Admin";
import BulkEditDocuments from "./components/Documents/BulkEdit";

const App = () => (
  <Admin
    layout={Layout}
    theme={lightTheme}
    dataProvider={dataProvider}
    authProvider={authProvider}
    requireAuth
  >
    <Resource name="documents" {...documents} icon={ArticleIcon} />
    <Resource
      name="metadata/keys"
      options={{ label: "Metadata" }}
      {...MetadataKeys}
      icon={TagIcon}
    />
    <Resource name="metadata/values" options={{ label: "metadata values" }} />
    <Resource
      name="processing/rules"
      options={{ label: "Processing" }}
      {...Rules}
      icon={ConstructionIcon}
    />

    <Resource name="user" />
    <Resource name="preferences" />
    <Resource name="admin" />
    <Resource name="admin/users" />
    <Resource name="admin/documents/processing" />
    <Resource name="documents/bulkEdit" create={<BulkEditDocuments/>} />


    <CustomRoutes>
      <Route
        key="preferences"
        path="/preferences"
        // @ts-ignore
        element={<ProfileEdit />}
      />
      ,
      <Route
        key="administrating"
        path={"/admin"}
        // @ts-ignore
        element={<AdminView />}
      />
      {/* <Route
        key="bulkEditDocuments"
        path={"/documents/bulkEdit/:ids"}
        // @ts-ignore
        element={<BulkEditDocuments/>}
      /> */}

    </CustomRoutes>
  </Admin>
);

export default App;
