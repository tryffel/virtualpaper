import {
  BooleanInput,
  NumberInput,
  SelectInput,
  SimpleForm,
  TextInput,
  useAuthProvider,
} from "react-admin";
import { useEffect, useState } from "react";

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
    : "Only admins can enable global";

  return (
    <SimpleForm>
      <TextInput source="name" />
      <SelectInput
        source={"property_type"}
        label={"Type"}
        choices={propertyTypes}
      />
      <BooleanInput
        label={"Property is shared"}
        source={"global"}
        disabled={!isAdmin}
        helperText={globalHelpText}
      />
      <BooleanInput
        label={"Property values are unique"}
        source={"unique"}
        helperText={
          "When unique, each property value can exist only for one document"
        }
      />
      <NumberInput source={"counter"} />
      <NumberInput source={"offset"} />
      <TextInput source={"prefix"} />
      <TextInput source={"mode"} />
      <BooleanInput
        source={"read_only"}
        helperText={
          "If enabled, assigned values cannot be edited after creation"
        }
      />
      <TextInput source={"date_format"} />
    </SimpleForm>
  );
};
