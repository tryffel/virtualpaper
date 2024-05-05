import {
  BooleanInput,
  DeleteWithConfirmButton,
  Form,
  NumberInput,
  SaveButton,
  SelectInput,
  ShowButton,
  TextInput,
  Toolbar,
  TopToolbar,
  useAuthProvider,
} from "react-admin";
import { useEffect, useState } from "react";
import { useWatch } from "react-hook-form";
import { Grid } from "@mui/material";

const propertyTypes = [
  {
    id: "text",
    name: "Text",
  },
  {
    id: "url",
    name: "Url",
  },
  {
    id: "id",
    name: "Id",
  },
  {
    id: "json",
    name: "JSON",
  },
  {
    id: "counter",
    name: "Counter",
  },
  {
    id: "int",
    name: "Number (no decimals)",
  },
  {
    id: "float",
    name: "Number (with decimals)",
  },
  {
    id: "boolean",
    name: "Boolean",
  },
  {
    id: "date",
    name: "Date",
  },
  {
    id: "user",
    name: "User",
  },
];

const ModeInput = () => {
  const selectedType = useWatch<{ property_type: string }>({
    name: "property_type",
  });
  return (
    <TextInput
      source={"mode"}
      label={"Id type"}
      disabled={selectedType !== "id"}
    />
  );
};

const DateFmtInput = () => {
  const selectedType = useWatch<{ property_type: string }>({
    name: "property_type",
  });
  return (
    <TextInput
      source={"date_format"}
      label={"Date format"}
      disabled={selectedType !== "date"}
    />
  );
};

const CounterInput = () => {
  const selectedType = useWatch<{ property_type: string }>({
    name: "property_type",
  });

  return (
    <NumberInput
      source={"counter"}
      label={"Counter value"}
      disabled={selectedType !== "counter"}
    />
  );
};

export const PropertyForm = () => {
  const auth = useAuthProvider();

  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    auth.getPermissions(null).then((perm) => {
      setIsAdmin(perm?.admin);
    });
  }, [auth]);

  const globalHelpText = isAdmin
    ? "Global property is shared and unique across all users"
    : "Only admins can manage global properties";

  return (
    <Form>
      <TopToolbar>
        <ShowButton />
      </TopToolbar>
      <Grid container gap={1} margin={1}>
        <Grid item xs={12} md={6}>
          <TextInput source="name" />
        </Grid>
        <Grid item xs={12} md={5}>
          <SelectInput
            source={"property_type"}
            label={"Type"}
            choices={propertyTypes}
            required
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <BooleanInput
            label={"Property is shared"}
            source={"global"}
            disabled={!isAdmin}
            helperText={globalHelpText}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <BooleanInput
            label={"Property values are unique"}
            source={"unique"}
            helperText={
              "When unique, each property value can exist only for one document"
            }
          />
        </Grid>
        <Grid item xs={12} sm={6} md={4}>
          <BooleanInput
            label={"Exclusive per document"}
            source={"exclusive"}
            helperText={
              "When enabled, each document can have property only once"
            }
          />
        </Grid>
        <Grid item xs={12} sm={6}>
          <BooleanInput
            source={"read_only"}
            helperText={
              "If enabled, assigned values cannot be edited after creation"
            }
          />
        </Grid>
        <Grid item xs={12} sm={6}>
          <CounterInput />
        </Grid>
        <Grid item xs={12} sm={6}>
          <TextInput source={"prefix"} />
        </Grid>
        <Grid item xs={12} sm={6}>
          <ModeInput />
        </Grid>

        <Grid item xs={12} sm={6}>
          <DateFmtInput />
        </Grid>
      </Grid>
      <Toolbar
        sx={{
          display: "flex",
          flexDirection: "row",
          justifyContent: "space-between",
        }}
      >
        <SaveButton />
        <DeleteWithConfirmButton />
      </Toolbar>
    </Form>
  );
};
