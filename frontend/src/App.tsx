import "./App.css";

import TagIcon from "@mui/icons-material/Tag";
import ConstructionIcon from "@mui/icons-material/Construction";
import ArticleIcon from "@mui/icons-material/Article";
import DeleteIcon from "@mui/icons-material/Delete";

import { Route } from "react-router-dom";

import { Admin, Resource, CustomRoutes } from "react-admin";
import { dataProvider } from "./api/dataProvider";
import authProvider from "./api/authProvider";
import documents from "./resources/Documents";

import { lightTheme, darkTheme } from "./theme";
import Layout from "./layout/Layout";
import MetadataKeys from "./resources/MetadataKeys";
import Rules from "./resources/Rules";

import { ProfileEdit } from "./resources/Preferences";
import AdminView from "./resources/Admin";
import BulkEditDocuments from "./resources/Documents/BulkEdit";
import { Dashboard } from "./resources/Dashboard";
import { ResetPassword } from "./resources/Public/ResetPassword";
import { ForgotPassword } from "./resources/Public/ForgotPassword";
import { AdminEditUser } from "./resources/Admin/UserEdit";
import { AdminCreateUser } from "./resources/Admin/UserCreate";
import { ConfirmAuthentication } from "./resources/Authentication/AuthConfirmationDialog";
import { DeletedDocumentList } from "./resources/Documents/Trashbin";
import { LoginPage } from "./resources/Public/Login";

const App = () => (
  <Admin
    layout={Layout}
    theme={lightTheme}
    darkTheme={darkTheme}
    dataProvider={dataProvider}
    authProvider={authProvider}
    dashboard={Dashboard}
    loginPage={<LoginPage />}
    disableTelemetry
  >
    <Resource name="documents" {...documents} icon={ArticleIcon} />
    <Resource
      name="documents/deleted"
      list={<DeletedDocumentList />}
      icon={DeleteIcon}
      options={{ label: "Trash bin" }}
    />
    <Resource
      name="metadata/keys"
      options={{ label: "Metadata" }}
      {...MetadataKeys}
      icon={TagIcon}
    />
    <Resource name="metadata/values" options={{ label: "metadata values" }} />
    <Resource
      name="processing/rules"
      options={{ label: "Processing rules" }}
      {...Rules}
      icon={ConstructionIcon}
    />

    <Resource name="user" />
    <Resource name="preferences" />
    <Resource name="admin" />
    <Resource
      name="admin/users"
      edit={<AdminEditUser />}
      create={<AdminCreateUser />}
      options={{ label: "Users" }}
    />
    <Resource name="admin/documents/processing" />
    <Resource name="documents/bulkEdit" create={<BulkEditDocuments />} />
    <Resource name="documents/linked" options={{ label: "Linked documents" }} />

    <Resource name="reset-password" />
    <Resource name="forgot-password" />

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
    </CustomRoutes>
    <CustomRoutes noLayout>
      <Route path={"/auth/reset-password"} element={<ResetPassword />} />
      <Route path={"/auth/forgot-password"} element={<ForgotPassword />} />
      <Route
        path={"/auth/confirm-authentication"}
        element={<ConfirmAuthentication />}
      />
    </CustomRoutes>
  </Admin>
);

export default App;
